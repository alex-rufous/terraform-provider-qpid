package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
)

func resourceGroupMember() *schema.Resource {
	return &schema.Resource{
		Create: createGroupMember,
		Read:   readGroupMember,
		Delete: deleteGroupMember,
		Update: updateGroupMember,
		Exists: existsGroupMember,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of GroupMember",
				Required:    true,
				ForceNew:    true,
			},
			"group_provider": {
				Type:        schema.TypeString,
				Description: "The name of group provider the group belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"group": {
				Type:        schema.TypeString,
				Description: "The name of group the member belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of group Mmember",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "ManagedGroupMember"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid group member type value : '%q'. Allowed values: \"managed\"", v))
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

func createGroupMember(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)
	attributes := toGroupMemberAttributes(d)
	groupProvider := d.Get("group_provider").(string)
	groupName := d.Get("group").(string)
	resp, err := client.CreateGroupMember(groupProvider, groupName, attributes)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetGroupMember(groupProvider, groupName, name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid group member'%s/%s': %s", groupProvider, name, resp.Status)
}

func toGroupMemberAttributes(d *schema.ResourceData) *map[string]interface{} {

	attributes := schemaToAttributes(d, resourceGroupMember().Schema, "group_provider", "group")
	return attributes
}

func readGroupMember(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)
	groupName := d.Get("group").(string)
	attributes, err := client.GetGroupMember(groupProvider, groupName, name)
	if err != nil {
		return err
	}
	return applyResourceAttributes(d, attributes, "group_provider", "group")
}

func existsGroupMember(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)
	groupName := d.Get("group").(string)
	attributes, err := client.GetGroupMember(groupProvider, groupName, name)
	if err != nil {
		return false, err
	}

	return len(*attributes) > 0, nil
}

func deleteGroupMember(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)
	groupName := d.Get("group").(string)
	resp, err := client.DeleteGroupMember(groupProvider, groupName, name)
	if err != nil {
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error deleting qpid group '%s' on node %s: %s", name, groupProvider, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateGroupMember(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	groupProvider := d.Get("group_provider").(string)
	groupName := d.Get("group").(string)
	attributes := toGroupAttributes(d)

	resp, err := client.UpdateGroupMember(groupProvider, groupName, name, attributes)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("error updating qpid group '%s' on node '%s': %s", name, groupProvider, resp.Status)
}
