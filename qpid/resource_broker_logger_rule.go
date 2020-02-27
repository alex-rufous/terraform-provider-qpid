package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"net/http"
)

func resourceBrokerLoggerRule() *schema.Resource {
	return &schema.Resource{
		Create: createBrokerLoggerRule,
		Read:   readBrokerLoggerRule,
		Delete: deleteBrokerLoggerRule,
		Update: updateBrokerLoggerRule,
		Exists: existsBrokerLoggerRule,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of rule",
				Required:    true,
				ForceNew:    true,
			},
			"broker_logger": {
				Type:        schema.TypeString,
				Description: "The name of broker logger this rule belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Type of rule",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "NameAndLevel" || value == "UserOrConnectionSpecific"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid broker logger rule type value : '%v'", v))
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
			"logger_name": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ForceNew: true,
			},
			"level": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "ALL" || value == "TRACE" || value == "DEBUG" || value == "INFO" || value == "WARN" || value == "ERROR" || value == "OFF"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid broker logger rule level value : '%v'", v))
					}

					return
				},
			},
			"connection_name": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
			},
			"remote_container_id": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Default:  nil,
				Optional: true,
			},
		},
	}
}

func createBrokerLoggerRule(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)
	attributes := toBrokerLoggerRuleAttributes(d)
	brokerLogger := d.Get("broker_logger").(string)
	resp, err := client.CreateBrokerLoggerRule(brokerLogger, attributes)
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetBrokerLoggerRule(brokerLogger, name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid broker logger rule '%s/%s': %s", brokerLogger, name, resp.Status)
}

func toBrokerLoggerRuleAttributes(d *schema.ResourceData) *map[string]interface{} {

	attributes := schemaToAttributes(d, resourceBrokerLoggerRule().Schema, "broker_logger")
	return attributes
}

func readBrokerLoggerRule(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	brokerLogger := d.Get("broker_logger").(string)

	attributes, err := client.GetBrokerLoggerRule(brokerLogger, name)
	if err != nil {
		return err
	}

	return applyResourceAttributes(d, attributes, "broker_logger")
}

func existsBrokerLoggerRule(d *schema.ResourceData, meta interface{}) (bool, error) {

	client := meta.(*Client)

	name := d.Get("name").(string)
	brokerLogger := d.Get("broker_logger").(string)
	attributes, err := client.GetBrokerLoggerRule(brokerLogger, name)
	if err != nil {
		return false, err
	}

	return len(*attributes) > 0, nil
}

func deleteBrokerLoggerRule(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	brokerLogger := d.Get("broker_logger").(string)

	resp, err := client.DeleteBrokerLoggerRule(brokerLogger, name)
	if err != nil {
		return nil
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("error deleting qpid broker logger rule '%s' on node %s: %s", name, brokerLogger, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateBrokerLoggerRule(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	name := d.Get("name").(string)
	brokerLogger := d.Get("broker_logger").(string)
	attributes := toBrokerLoggerRuleAttributes(d)

	resp, err := client.UpdateBrokerLoggerRule(brokerLogger, name, attributes)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	m, _ := getErrorResponse(resp)
	return fmt.Errorf("error updating qpid broker logger rule '%s' on node '%s': %s, %v", name, brokerLogger, resp.Status, m)
}
