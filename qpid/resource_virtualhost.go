package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
)

func resourceVirtualHost() *schema.Resource {
	return &schema.Resource{
		Create: createVirtualHost,
		Read:   readVirtualHost,
		Delete: deleteVirtualHost,
		Update: updateVirtualHost,
		Exists: existsVirtualHost,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of Virtual Host",
				Required:    true,
				ForceNew:    true,
			},
			"virtual_host_node": {
				Type:        schema.TypeString,
				Description: "The name of Virtual Host Node",
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of Virtual Host",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "ProvidedStore" || value == "BDB" || value == "DERBY" || value == "JDBC"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid virtual host type value : '%v'. Allowed values: \"ProvidedStore\", \"BDB\", \"DERBY\", \"JDBC\"", v))
					}

					return
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
				Default:  nil,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {

					// ignore when broker reports that object is durable but attribute was not set explicitly in configuration
					keyValue, keySet := d.GetOk(k)

					log.Printf("durable is set %v old '%s', new '%s', value %v", keySet, old, new, keyValue)
					return !keySet && (old == "true" || new == "true")
				},
			},

			"context": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// only applicable for types BDB, DERBY
			"store_path": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password", "table_name_prefix"},
			},

			// only applicable for type JDBC
			"connection_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"store_path", "local_transaction_synchronization_policy", "remote_transaction_synchronization_policy", "coalescing_sync", "durability", "store_underfull_size", "store_overfull_size"},
			},
			// only applicable for type JDBC
			"connection_pool_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"store_path", "local_transaction_synchronization_policy", "remote_transaction_synchronization_policy", "coalescing_sync", "durability", "store_underfull_size", "store_overfull_size"},
			},
			// only applicable for type JDBC
			"username": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"store_path", "local_transaction_synchronization_policy", "remote_transaction_synchronization_policy", "coalescing_sync", "durability", "store_underfull_size", "store_overfull_size"},
			},
			// only applicable for type JDBC
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				Sensitive:     true,
				ConflictsWith: []string{"store_path", "local_transaction_synchronization_policy", "remote_transaction_synchronization_policy", "coalescing_sync", "durability", "store_underfull_size", "store_overfull_size"},
			},
			// only applicable for type JDBC
			"table_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"store_path", "local_transaction_synchronization_policy", "remote_transaction_synchronization_policy", "coalescing_sync", "durability", "store_underfull_size", "store_overfull_size"},
			},

			// only applicable for type BDB_HA
			"local_transaction_synchronization_policy": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},
			// only applicable for type BDB_HA
			"remote_transaction_synchronization_policy": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},
			// only applicable for type BDB_HA
			"coalescing_sync": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},
			// only applicable for type BDB_HA
			"durability": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},
			// only applicable for types BDB_HA, BDB and DERBY
			"store_underfull_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},
			// only applicable for types BDB_HA, BDB and DERBY
			"store_overfull_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},
			// common
			"statistics_reporting_period": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"store_transaction_idle_timeout_close": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"store_transaction_idle_timeout_warn": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"store_transaction_open_timeout_close": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"store_transaction_open_timeout_warn": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"housekeeping_check_period": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"housekeeping_thread_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"connection_thread_pool_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"number_of_selectors": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},
			"enabled_connection_validators": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},
			"disabled_connection_validators": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},
			"global_address_domains": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},
			"node_auto_creation_policy": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pattern": {
							Type:     schema.TypeString,
							Required: true,
						},
						"created_on_publish": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"created_on_consume": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"node_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"attributes": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func createVirtualHost(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)
	attributes := toVirtualHostAttributes(d)
	node := d.Get("virtual_host_node").(string)
	resp, err := client.CreateVirtualHost(node, attributes)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetVirtualHost(node, name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid virtual host '%s/%s': %s", node, name, resp.Status)
}

func toVirtualHostAttributes(d *schema.ResourceData) *map[string]interface{} {

	attributes := schemaToAttributes(d, resourceVirtualHost().Schema, "virtual_host_node", "node_auto_creation_policy")

	value, exists := d.GetOk("node_auto_creation_policy")
	if exists {
		var val, expected = value.(*schema.Set)
		if expected && val != nil {
			var items = val.List()
			for i, v := range items {
				p := v.(map[string]interface{})
				items[i] = *createMapWithKeysInCameCase(&p)
			}
			(*attributes)["nodeAutoCreationPolicies"] = items
		}
	} else {
		oldValue, newValue := d.GetChange("node_auto_creation_policy")
		if fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", newValue) {
			(*attributes)["nodeAutoCreationPolicies"] = nil
		}
	}

	return attributes
}

func readVirtualHost(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	node := d.Get("virtual_host_node").(string)

	attributes, err := client.GetVirtualHost(node, name)
	if err != nil {
		return err
	}

	err = applyResourceAttributes(d, attributes, "node_auto_creation_policy", "virtual_host_node")
	if err != nil {
		return err
	}

	keyValue, keySet := d.GetOk("node_auto_creation_policy")
	value, attributeSet := (*attributes)["nodeAutoCreationPolicies"]
	if keySet || attributeSet {
		if value != nil {
			s := keyValue.(*schema.Set)
			if s != nil {
				for _, v := range value.([]interface{}) {
					p := v.(map[string]interface{})
					s.Add(*createMapWithKeysUnderscored(&p))
				}
			}
			value = s
		}

		err = d.Set("node_auto_creation_policy", value)
		if err != nil {
			return err
		}
	}

	return nil
}

func existsVirtualHost(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	name := d.Get("name").(string)
	node := d.Get("virtual_host_node").(string)
	attributes, err := client.GetVirtualHost(node, name)
	if err != nil {
		return false, err
	}

	return len(*attributes) > 0, nil
}

func deleteVirtualHost(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	node := d.Get("virtual_host_node").(string)

	resp, err := client.DeleteVirtualHost(node, name)
	if err != nil {
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error deleting qpid virtual host '%s' on node %s: %s", name, node, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateVirtualHost(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	node := d.Get("virtual_host_node").(string)
	attributes := toVirtualHostAttributes(d)

	resp, err := client.UpdateVirtualHost(node, name, attributes)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("error updating qpid virtual host '%s' on node '%s': %s", name, node, resp.Status)
}
