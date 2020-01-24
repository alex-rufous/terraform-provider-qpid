package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
	"strings"
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

			"parents": {
				Type:        schema.TypeList,
				Description: "Parents of Exchange",
				Required:    true,
				ForceNew:    true,
				MaxItems:    2,
				MinItems:    2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
				Default:  nil,
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
							Default: nil,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
				Default:  nil,
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
	items := d.Get("parents").([]interface{})
	var parents = *convertToArrayOfStrings(&items)

	if len(parents) != 2 {
		return fmt.Errorf("unexpected exchange parents: %s", strings.Join(parents, "/"))
	}

	resp, err := client.CreateExchange(parents[0], parents[1], attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetExchange(parents[0], parents[1], name)
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
	attributes := make(map[string]interface{})
	schemaMap := resourceExchange().Schema
	for key := range schemaMap {
		var value interface{}
		value, exists := d.GetOk(key)
		if key != "parents" && exists {
			if key == "alternate_binding" {
				val, expected := value.([]interface{})
				if expected && len(val) == 1 {
					i := val[0].(map[string]interface{})
					value = createMapWithKeysInCameCase(&i)
				}
			}
			attributes[convertToCamelCase(key)] = value
		}
	}
	return &attributes
}

func readExchange(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	items := d.Get("parents").([]interface{})
	var parents = *convertToArrayOfStrings(&items)
	name := d.Get("name").(string)
	attributes, err := client.GetExchange(parents[0], parents[1], name)
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

		if key!="parents" && ( keySet || attributeSet ){
			isString := false
			if value != nil {
				_, isString =  value.(string)
			}
			log.Printf("exchange attribute: %s=%v, is string: %v", key, value, isString)

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

	items := d.Get("parents").([]interface{})
	var parents = *convertToArrayOfStrings(&items)
	name := d.Get("name").(string)
	attributes, err := client.GetExchange(parents[0], parents[1], name)
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

	items := d.Get("parents").([]interface{})
	var parents = *convertToArrayOfStrings(&items)
	name := d.Get("name").(string)
	resp, err := client.DeleteExchange(parents[0], parents[1], name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound{
		return fmt.Errorf("error deleting qpid exchange '%s' on virtual host %s/%s: %d", name, parents[0], parents[1], resp.StatusCode)
	}
	d.SetId("")
	return nil
}

func updateExchange(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	items := d.Get("parents").([]interface{})
	var parents = *convertToArrayOfStrings(&items)
	name := d.Get("name").(string)

	attributes := toExchangeAttributes(d)
	resp, err := client.UpdateExchange(parents[0], parents[1], name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid exchange '%s' on virtual host '%s/%s' does not exist", name, parents[0], parents[1])
	}

	return fmt.Errorf("error updating qpid exchange '%s' on virtua host '%s/%s': %s", name, parents[0], parents[1], resp.Status)
}
