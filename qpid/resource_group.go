package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: createGroup,
		Read:   readGroup,
		Delete: deleteGroup,
		Update: updateGroup,
		Exists: existsGroup,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of Group",
				Required:    true,
				ForceNew:    true,
			},
			"group_provider": {
				Type:        schema.TypeString,
				Description: "The name of group provider the group belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of Group",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "ManagedGroup"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid group type value : '%v'", v))
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
		},
	}
}

func createGroup(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)
	attributes := toGroupAttributes(d)
	groupProvider := d.Get("group_provider").(string)
	resp, err := client.CreateGroup(groupProvider, attributes)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetGroup(groupProvider, name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid group '%s/%s': %s", groupProvider, name, resp.Status)
}

func toGroupAttributes(d *schema.ResourceData) *map[string]interface{} {

	attributes := schemaToAttributes(d, resourceGroup().Schema, "group_provider")
	return attributes
}

func readGroup(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)

	attributes, err := client.GetGroup(groupProvider, name)
	if err != nil {
		return err
	}

	return applyResourceAttributes(d, attributes, "group_provider")
}

func existsGroup(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)
	attributes, err := client.GetGroup(groupProvider, name)
	if err != nil {
		return false, err
	}

	return len(*attributes) > 0, nil
}

func deleteGroup(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)

	resp, err := client.DeleteGroup(groupProvider, name)
	if err != nil {
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error deleting qpid group '%s' on node %s: %s", name, groupProvider, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateGroup(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)
	attributes := toGroupAttributes(d)

	resp, err := client.UpdateGroup(groupProvider, name, attributes)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("error updating qpid group '%s' on node '%s': %s", name, groupProvider, resp.Status)
}
