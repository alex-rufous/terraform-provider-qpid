package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceExchange() *schema.Resource {
	return &schema.Resource{
		Create: createExchange,
		Read:   readExchange,
		Delete: deleteExchange,
		Update: updateExchange,
		Exists: existsExchange,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of Exchange",
				Required:    true,
				ForceNew:    true,
			},

			"virtual_host_node": {
				Type:        schema.TypeString,
				Description: "The name of Virtual Host Node",
				Required:    true,
				ForceNew:    true,
			},

			"virtual_host": {
				Type:        schema.TypeString,
				Description: "The name of Virtual Host",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of Exchange",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "direct" || value == "topic" || value == "fanout" || value == "headers"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid exchange type value : '%q'. Allowed values: \"direct\", \"topic\", \"fanout\", \"headers\"", v))
					}

					return
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					_, keySet := d.GetOk(k)
					return !keySet && (old == "direct" || new == "direct")
				},
			},
			"description": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
			},
			"durable": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					_, keySet := d.GetOk(k)
					return !keySet && (old == "true" || new == "true")
				},
			},

			"context": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},

			"alternate_binding": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination": {
							Type:     schema.TypeString,
							Required: true,
						},
						"attributes": {
							Type:     schema.TypeMap,
							Optional: true,
							Default:  nil,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				Default: nil,
			},

			"unroutable_message_behaviour": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "REJECT" || value == "DISCARD"
					if !valid {
						errors = append(errors, fmt.Errorf("invalid exchange exclusivity policy : '%s'", v))
					}
					return
				},
			},
		},
	}
}

func createExchange(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	attributes := toExchangeAttributes(d)
	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}

	resp, err := client.CreateExchange(node.(string), host.(string), attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetExchange(node.(string), host.(string), name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid exchange'%s': %s", name, resp.Status)
}

func toExchangeAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceExchange().Schema, "virtual_host_node", "virtual_host")
	alternateBinding, alternateBindingSet := (*attributes)["alternateBinding"]
	if alternateBindingSet && alternateBinding != nil {
		val, expected := alternateBinding.([]interface{})
		if expected && len(val) == 1 {
			i := val[0].(map[string]interface{})
			binding := createMapWithKeysInCameCase(&i)
			(*attributes)["alternateBinding"] = binding
		} else {
			delete(*attributes, "alternateBinding")
		}
	}
	return attributes
}

func readExchange(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}
	name := d.Get("name").(string)
	attributes, err := client.GetExchange(node.(string), host.(string), name)
	if err != nil {
		return err
	}

	if len(*attributes) == 0 {
		return nil
	}

	schemaMap := resourceExchange().Schema
	for key, v := range schemaMap {
		_, keySet := d.GetOk(key)
		keyCamelCased := convertToCamelCase(key)
		value, attributeSet := (*attributes)[keyCamelCased]

		if key != "virtual_host_node" && key != "virtual_host" && (keySet || attributeSet) {

			if key == "alternate_binding" {
				val := value.(map[string]interface{})
				value = []interface{}{createMapWithKeysUnderscored(&val)}
			}

			value, err = convertIfValueIsStringWhenPrimitiveIsExpected(value, v.Type)
			if err != nil {
				return err
			}
			err = d.Set(key, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func existsExchange(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return false, fmt.Errorf("virtual_host_node and virtual_host are not set")
	}
	name := d.Get("name").(string)
	attributes, err := client.GetExchange(node.(string), host.(string), name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}

	return true, nil
}

func deleteExchange(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}
	name := d.Get("name").(string)
	resp, err := client.DeleteExchange(node.(string), host.(string), name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid exchange '%s' on virtual host %s/%s: %d", name, node, host, resp.StatusCode)
	}
	d.SetId("")
	return nil
}

func updateExchange(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}
	name := d.Get("name").(string)

	attributes := toExchangeAttributes(d)
	resp, err := client.UpdateExchange(node.(string), host.(string), name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid exchange '%s' on virtual host '%s/%s' does not exist", name, node, host)
	}

	return fmt.Errorf("error updating qpid exchange '%s' on virtua host '%s/%s': %s", name, node, host, resp.Status)
}
