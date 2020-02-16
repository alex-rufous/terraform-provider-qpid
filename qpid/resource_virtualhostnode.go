package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
	"net/http"
)

func resourceVirtualHostNode() *schema.Resource {
	return &schema.Resource{
		Create: createVirtualHostNode,
		Read:   readVirtualHostNode,
		Delete: deleteVirtualHostNode,
		Update: updateVirtualHostNode,
		Exists: existsVirtualHostNode,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of Virtual Host Node",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of Virtual Host Node",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "JSON" || value == "BDB" || value == "DERBY" || value == "BDB_HA" || value == "JDBC"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid virtual host node type value : '%q'. Allowed values: \"JSON\", \"BDB\", \"DERBY\", \"BDB_HA\", \"JDBC\"", v))
					}

					return
				},
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  nil,
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
					return !(!keySet && old == "true")
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

			"default_virtual_host_node": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"virtual_host_initial_configuration": {
				Type:         schema.TypeString,
				Default:      nil,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validation.ValidateJsonString,
				// virtualHostInitialConfiguration only used once on creation and changed on broker side to an json object after that
				// thus, the real value from the broker would be an empty json object '{}'
				// disregard any difference when old value is  '{}'
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "{}"
				},
			},

			// only applicable for types BDB, BDB_HA, JSON and DERBY
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
				Default:       nil,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"store_path", "group_name", "helper_address", "designated_primary", "priority", "quorum_override", "helper_node_name", "permitted_nodes"},
			},

			// only applicable for type JDBC
			"connection_pool_type": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"store_path", "group_name", "helper_address", "designated_primary", "priority", "quorum_override", "helper_node_name", "permitted_nodes"},
			},

			// only applicable for type JDBC
			"username": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"store_path", "group_name", "helper_address", "designated_primary", "priority", "quorum_override", "helper_node_name", "permitted_nodes"},
			},

			// only applicable for type JDBC
			"password": {
				Type:          schema.TypeString,
				Default:       nil,
				Sensitive:     true,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"store_path", "group_name", "helper_address", "designated_primary", "priority", "quorum_override", "helper_node_name", "permitted_nodes"},
			},

			// only applicable for type JDBC
			"table_name_prefix": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"store_path", "group_name", "helper_address", "designated_primary", "priority", "quorum_override", "helper_node_name", "permitted_nodes"},
			},

			// only applicable for type BDB_HA
			"group_name": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},

			// only applicable for type BDB_HA
			"address": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},

			// only applicable for type BDB_HA
			"helper_address": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},

			// only applicable for type BDB_HA
			"designated_primary": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
				Default:       nil,
			},

			// only applicable for type BDB_HA
			"priority": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},

			// only applicable for type BDB_HA
			"quorum_override": {
				Type:          schema.TypeInt,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},

			// only applicable for type BDB_HA
			"helper_node_name": {
				Type:          schema.TypeString,
				Default:       nil,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
			},

			// only applicable for type BDB_HA
			"permitted_nodes": {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      false,
				ConflictsWith: []string{"connection_url", "connection_pool_type", "username", "password"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func createVirtualHostNode(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	attributes := toVirtualHostNodeAttributes(d)
	resp, err := client.CreateVirtualHostNode(attributes)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			attributes, err = client.GetVirtualHostNode(name)
			if err != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid virtual host node '%s': %s", name, resp.Status)
}

func toVirtualHostNodeAttributes(d *schema.ResourceData) *map[string]interface{} {
	return schemaToAttributes(d, resourceVirtualHostNode().Schema)
}

func readVirtualHostNode(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetVirtualHostNode(name)
	if err != nil {
		return err
	}

	if len(*attributes) == 0 {
		d.SetId("")
		return nil
	}

	schemaMap := resourceVirtualHostNode().Schema
	for key := range schemaMap {
		_, keySet := d.GetOk(key)
		keyCamelCased := convertToCamelCase(key)
		value, attributeSet := (*attributes)[keyCamelCased]

		if keySet || attributeSet {

			if key == "permitted_nodes" && value != nil {
				val, expected := value.([]interface{})
				if expected {
					value = convertToArrayOfStrings(&val)
				} else {
					return fmt.Errorf("unexpected value set for %s: %v", key, value)
				}
			}

			err = d.Set(key, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func existsVirtualHostNode(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetVirtualHostNode(name)
	if err != nil {
		return false, err
	}
	return len(*attributes) != 0, nil
}

func deleteVirtualHostNode(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	resp, err := client.DeleteVirtualHostNode(name)
	if err != nil {
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error deleting qpid virtual host node %s: %d", name, resp.StatusCode)
	}
	d.SetId("")
	return nil
}

func updateVirtualHostNode(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toVirtualHostNodeAttributes(d)

	resp, err := client.UpdateVirtualHostNode(name, attributes)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("error updating qpid virtual host node '%s': %s", name, resp.Status)
}
