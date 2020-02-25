package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourcePort() *schema.Resource {

	return &schema.Resource{
		Create: createPort,
		Read:   readPort,
		Delete: deletePort,
		Update: updatePort,
		Exists: existsPort,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of port",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of port",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "AMQP" || value == "HTTP"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid port type value : '%q'", v))
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

			"port": {
				Type:     schema.TypeInt,
				Required: true,
				Default:  nil,
			},

			"allow_confidential_operations_on_insecure_channels": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  nil,
			},

			"protocols": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"transports": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"binding_address": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
			},

			"authentication_provider": {
				Type:     schema.TypeString,
				Default:  nil,
				Required: true,
			},

			"key_store": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
			},

			"trust_stores": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"need_client_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  nil,
			},

			"want_client_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  nil,
			},

			"client_cert_recorder": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// AMQP
			"tcp_to_delay": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       nil,
				ConflictsWith: []string{"thread_pool_maximum", "thread_pool_minimum", "manage_broker_on_no_alias_match"},
			},

			"thread_pool_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				Default:       nil,
				ConflictsWith: []string{"thread_pool_maximum", "thread_pool_minimum", "manage_broker_on_no_alias_match"},
			},

			"number_of_selectors": {
				Type:          schema.TypeInt,
				Optional:      true,
				Default:       nil,
				ConflictsWith: []string{"thread_pool_maximum", "thread_pool_minimum", "manage_broker_on_no_alias_match"},
			},

			"max_open_connections": {
				Type:          schema.TypeInt,
				Optional:      true,
				Default:       nil,
				ConflictsWith: []string{"thread_pool_maximum", "thread_pool_minimum", "manage_broker_on_no_alias_match"},
			},

			// HTTP
			"thread_pool_maximum": {
				Type:          schema.TypeInt,
				Optional:      true,
				Default:       nil,
				ConflictsWith: []string{"tcp_to_delay", "thread_pool_size", "number_of_selectors", "max_open_connections"},
			},

			"thread_pool_minimum": {
				Type:          schema.TypeInt,
				Optional:      true,
				Default:       nil,
				ConflictsWith: []string{"tcp_to_delay", "thread_pool_size", "number_of_selectors", "max_open_connections"},
			},

			"manage_broker_on_no_alias_match": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       nil,
				ConflictsWith: []string{"tcp_to_delay", "thread_pool_size", "number_of_selectors", "max_open_connections"},
			},
		},
	}
}

func createPort(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	attributes := toPortAttributes(d)

	resp, err := client.CreatePort(attributes)

	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetPort(name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	m, _ := getErrorResponse(resp)

	return fmt.Errorf("error creating qpid port'%s': %s, %v", name, resp.Status, m)
}

func toPortAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourcePort().Schema)
	(*attributes)["port"] = d.Get("port")
	return attributes
}

func readPort(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetPort(name)
	if err != nil {
		return err
	}

	if len(*attributes) == 0 {
		return nil
	}

	return applyResourceAttributes(d, attributes)
}

func existsPort(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	name := d.Get("name").(string)
	attributes, err := client.GetPort(name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}
	return true, nil
}

func deletePort(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	resp, err := client.DeletePort(name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid port '%s': %s", name, resp.Status)
	}
	d.SetId("")
	return nil
}

func updatePort(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toPortAttributes(d)

	resp, err := client.UpdatePort(name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid  port '%s' does not exist", name)
	}

	return fmt.Errorf("error updating qpid port '%s': %s", name, resp.Status)
}
