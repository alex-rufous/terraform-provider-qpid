package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Create: createUser,
		Read:   readUser,
		Delete: deleteUser,
		Update: updateUser,
		Exists: existsUser,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of User",
				Required:    true,
				ForceNew:    true,
			},
			"authentication_provider": {
				Type:        schema.TypeString,
				Description: "The name of authentication provider user belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of User",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "managed"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid user type value : '%v'. Allowed values: \"managed\"", v))
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

			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return true
				},
			},
		},
	}
}

func createUser(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)
	attributes := toUserAttributes(d)
	authenticationProvider := d.Get("authentication_provider").(string)
	resp, err := client.CreateUser(authenticationProvider, attributes)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetUser(authenticationProvider, name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid user '%s/%s': %s", authenticationProvider, name, resp.Status)
}

func toUserAttributes(d *schema.ResourceData) *map[string]interface{} {

	attributes := schemaToAttributes(d, resourceUser().Schema, "authentication_provider")
	return attributes
}

func readUser(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	authenticationProvider := d.Get("authentication_provider").(string)

	attributes, err := client.GetUser(authenticationProvider, name)
	if err != nil {
		return err
	}

	return applyResourceAttributes(d, attributes, "authentication_provider")
}

func existsUser(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	name := d.Get("name").(string)
	authenticationProvider := d.Get("authentication_provider").(string)
	attributes, err := client.GetUser(authenticationProvider, name)
	if err != nil {
		return false, err
	}

	return len(*attributes) > 0, nil
}

func deleteUser(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	authenticationProvider := d.Get("authentication_provider").(string)

	resp, err := client.DeleteUser(authenticationProvider, name)
	if err != nil {
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error deleting qpid user '%s' on node %s: %s", name, authenticationProvider, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateUser(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	authenticationProvider := d.Get("authentication_provider").(string)
	attributes := toUserAttributes(d)

	resp, err := client.UpdateUser(authenticationProvider, name, attributes)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("error updating qpid user '%s' on node '%s': %s", name, authenticationProvider, resp.Status)
}
