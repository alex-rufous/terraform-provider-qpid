package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
)

func resourceVirtualHostAlias() *schema.Resource {
	return &schema.Resource{
		Create: createVirtualHostAlias,
		Read:   readVirtualHostAlias,
		Delete: deleteVirtualHostAlias,
		Update: updateVirtualHostAlias,
		Exists: existsVirtualHostAlias,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of alias",
				Required:    true,
				ForceNew:    true,
			},

			"port": {
				Type:        schema.TypeString,
				Description: "The name of port this alias belongs to",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of alias",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "hostnameAlias" || value == "patternMatchingAlias" || value == "systemAddressAlias" || value == "nameAlias" || value == "defaultAlias"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid alias type value : '%q'. Allowed values: \"managed\"", v))
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

			"priority": {
				Type:     schema.TypeInt,
				Default:  nil,
				Optional: true,
			},

			"pattern": {
				Type:        schema.TypeString,
				Description: "Pattern for matching virtual host name. Applicable to types: patternMatchingAlias and systemAddressAlias",
				Optional:    true,
				Default:     nil,
			},

			"system_address_space": {
				Type:        schema.TypeString,
				Description: "System address space. Applicable only to systemAddressAlias",
				Optional:    true,
				Default:     nil,
			},
		},
	}
}

func createVirtualHostAlias(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)
	attributes := toVirtualHostAliasAttributes(d)
	port := d.Get("port").(string)
	resp, err := client.CreateVirtualHostAlias(port, attributes)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetVirtualHostAlias(port, name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid alias '%s/%s': %s", port, name, resp.Status)
}

func toVirtualHostAliasAttributes(d *schema.ResourceData) *map[string]interface{} {

	attributes := schemaToAttributes(d, resourceVirtualHostAlias().Schema, "port")
	return attributes
}

func readVirtualHostAlias(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	port := d.Get("port").(string)

	attributes, err := client.GetVirtualHostAlias(port, name)
	if err != nil {
		return err
	}

	return applyResourceAttributes(d, attributes, "port")
}

func existsVirtualHostAlias(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	name := d.Get("name").(string)
	port := d.Get("port").(string)
	attributes, err := client.GetVirtualHostAlias(port, name)
	if err != nil {
		return false, err
	}

	return len(*attributes) > 0, nil
}

func deleteVirtualHostAlias(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	port := d.Get("port").(string)

	resp, err := client.DeleteVirtualHostAlias(port, name)
	if err != nil {
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error deleting qpid alias '%s' on port %s: %s", name, port, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateVirtualHostAlias(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	port := d.Get("port").(string)
	attributes := toVirtualHostAliasAttributes(d)

	resp, err := client.UpdateVirtualHostAlias(port, name, attributes)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("error updating qpid alias '%s' on port '%s': %s", name, port, resp.Status)
}
