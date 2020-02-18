package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceGroupProvider() *schema.Resource {

	return &schema.Resource{
		Create: createGroupProvider,
		Read:   readGroupProvider,
		Delete: deleteGroupProvider,
		Update: updateGroupProvider,
		Exists: existsGroupProvider,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			"name": {
				Type:        schema.TypeString,
				Description: "Name of group provider",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of group provider",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "CloudFoundryDashboardManagement" || value == "GroupFile" ||
						value == "ManagedGroupProvider"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid group provider type value : '%q'", v))
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

			// GroupFile provider fields
			"path": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"cloud_foundry_endpoint_u_r_i", "trust_store", "service_to_management_group_mapping"},
			},

			// CloudFoundryDashboardManagement provider fields
			"cloud_foundry_endpoint_u_r_i": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"path"},
			},

			"trust_store": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"path"},
			},

			"service_to_management_group_mapping": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"path"},
			},
		},
	}
}

func createGroupProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	attributes := toGroupProviderAttributes(d)
	resp, err := client.CreateGroupProvider(attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetGroupProvider(name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid group provider'%s': %s", name, resp.Status)
}

func toGroupProviderAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceGroupProvider().Schema)
	return attributes
}

func readGroupProvider(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetGroupProvider(name)
	if err != nil {
		return err
	}

	if len(*attributes) == 0 {
		return nil
	}

	schemaMap := resourceGroupProvider().Schema
	for key, v := range schemaMap {
		_, keySet := d.GetOk(key)
		keyCamelCased := convertToCamelCase(key)
		value, attributeSet := (*attributes)[keyCamelCased]

		if keySet || attributeSet {
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

func existsGroupProvider(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	name := d.Get("name").(string)
	attributes, err := client.GetGroupProvider(name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}
	return true, nil
}

func deleteGroupProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	resp, err := client.DeleteGroupProvider(name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid group provider '%s': %s", name, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateGroupProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toGroupProviderAttributes(d)

	resp, err := client.UpdateGroupProvider(name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid  group provider '%s' does not exist", name)
	}

	return fmt.Errorf("error updating qpid group provider '%s': %s", name, resp.Status)
}
