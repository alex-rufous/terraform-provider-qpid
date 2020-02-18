package qpid

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"net/http"
)

func resourceQueue() *schema.Resource {
	return &schema.Resource{
		Create: createQueue,
		Read:   readQueue,
		Delete: deleteQueue,
		Update: updateQueue,
		Exists: existsQueue,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of Queue",
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
				Description: "Type of Queue",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "standard" || value == "priority" || value == "sorted" || value == "lvq"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid queue type value : '%q'. Allowed values: \"standard\", \"priority\", \"sorted\", \"lvq\"", v))
					}

					return
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					_, keySet := d.GetOk(k)
					return !keySet && (old == "standard" || new == "standard")
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
				ForceNew: true,
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

			"exclusive": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "NONE" || value == "SESSION" || value == "CONNECTION" || value == "CONTAINER" || value == "PRINCIPAL" || value == "LINK" || value == "SHARED_SUBSCRIPTION"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid queue exclusivity policy : '%s'", v))
					}
					return
				},
			},

			"ensure_nondestructive_consumers": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"no_local": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"message_group_key_override": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"message_group_default_group": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"maximum_distinct_groups": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"message_group_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "NONE" || value == "STANDARD" || value == "SHARED_GROUPS"
					if !valid {
						errors = append(errors, fmt.Errorf("invalid message group type : '%s'", v))
					}
					return
				},
			},

			"maximum_delivery_attempts": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"alert_threshold_message_age": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"alert_threshold_message_size": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"alert_threshold_queue_depth_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"alert_threshold_queue_depth_messages": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"alert_repeat_gap": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"message_durability": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "DEFAULT" || value == "ALWAYS" || value == "NEVER"
					if !valid {
						errors = append(errors, fmt.Errorf("invalid message durability : '%s'", v))
					}
					return
				},
			},

			"minimum_message_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"maximum_message_ttl": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"default_filters": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.ValidateJsonString,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Default:          nil,
			},

			"hold_on_publish_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"maximum_queue_depth_bytes": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"overflow_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "NONE" || value == "RING" || value == "PRODUCER_FLOW_CONTROL" || value == "FLOW_TO_DISK" || value == "REJECT"
					if !valid {
						errors = append(errors, fmt.Errorf("invalid overflow policy : '%s'", v))
					}
					return
				},
			},

			"expiry_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "DELETE" || value == "ROUTE_TO_ALTERNATE"
					if !valid {
						errors = append(errors, fmt.Errorf("invalid expiry policy : '%s'", v))
					}
					return
				},
			},

			"lvq_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"sort_key", "priorities"},
			},

			"priorities": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"sort_key", "lvq_key"},
			},

			"sort_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"priorities", "lvq_key"},
			},
		},
	}
}

func createQueue(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	attributes := toQueueAttributes(d)
	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}
	resp, err := client.CreateQueue(node.(string), host.(string), attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetQueue(node.(string), host.(string), name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid queue'%s': %s", name, resp.Status)
}

func toQueueAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceQueue().Schema, "virtual_host_node", "virtual_host")
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

func readQueue(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}

	name := d.Get("name").(string)
	attributes, err := client.GetQueue(node.(string), host.(string), name)
	if err != nil {
		return err
	}

	if len(*attributes) == 0 {
		return nil
	}

	schemaMap := resourceQueue().Schema
	for key, v := range schemaMap {
		_, keySet := d.GetOk(key)
		keyCamelCased := convertToCamelCase(key)
		value, attributeSet := (*attributes)[keyCamelCased]

		if key != "virtual_host_node" && key != "virtual_host" && (keySet || attributeSet) {
			isString := false
			if value != nil {
				_, isString = value.(string)
			}

			if key == "alternate_binding" {
				val := value.(map[string]interface{})
				value = []interface{}{createMapWithKeysUnderscored(&val)}
			}
			if key == "default_filters" && value != nil && !isString {
				jsonData, err := json.Marshal(value)
				if err != nil {
					return err
				}
				value = string(jsonData)
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

func existsQueue(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return false, fmt.Errorf("virtual_host_node and virtual_host are not set")
	}

	name := d.Get("name").(string)
	attributes, err := client.GetQueue(node.(string), host.(string), name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}

	return true, nil
}

func deleteQueue(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}
	name := d.Get("name").(string)
	resp, err := client.DeleteQueue(node.(string), host.(string), name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid queue '%s' on virtual host %s/%s: %d", name, node, host, resp.StatusCode)
	}
	d.SetId("")
	return nil
}

func updateQueue(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	node := d.Get("virtual_host_node")
	host := d.Get("virtual_host")

	if host == nil || node == nil {
		return fmt.Errorf("virtual_host_node and virtual_host are not set")
	}

	name := d.Get("name").(string)

	attributes := toQueueAttributes(d)
	resp, err := client.UpdateQueue(node.(string), host.(string), name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid queue '%s' on virtual host '%s/%s' does not exist", name, node, host)
	}

	return fmt.Errorf("error updating qpid queue '%s' on virtua host '%s/%s': %s", name, node, host, resp.Status)
}
