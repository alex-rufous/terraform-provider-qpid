package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceAccessControlProvider() *schema.Resource {

	return &schema.Resource{
		Create: createAccessControlProvider,
		Read:   readAccessControlProvider,
		Delete: deleteAccessControlProvider,
		Update: updateAccessControlProvider,
		Exists: existsAccessControlProvider,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of access control provider",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of access control provider",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "AllowAll" || value == "AclFile" || value == "RuleBased"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid access control provider type value : '%v'", v))
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

			"priority": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  nil,
			},

			// RuleBased
			"rule": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identity": {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(string)
								valid := value == "ALL" || value == "VIRTUALHOSTNODE" || value == "VIRTUALHOST" ||
									value == "MANAGEMENT" || value == "QUEUE" || value == "EXCHANGE" ||
									value == "USER" || value == "GROUP" || value == "BROKER" || value == "METHOD"

								if !valid {
									errors = append(errors, fmt.Errorf("invalid object type value : '%v'", v))
								}

								return
							},
						},
						"operation": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(string)
								valid := value == "ALL" || value == "CONSUME" || value == "PUBLISH" ||
									value == "CREATE" || value == "UPDATE" || value == "DELETE" || value == "ACCESS" ||
									value == "CONFIGURE" || value == "BIND" || value == "UNBIND" || value == "INVOKE" ||
									value == "PURGE" || value == "ACCESS_LOGS" || value == "SHUTDOWN"

								if !valid {
									errors = append(errors, fmt.Errorf("invalid operation value : '%v'", v))
								}

								return
							},
						},
						"attributes": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"outcome": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								value := v.(string)
								valid := value == "ALLOW" || value == "ALLOW_LOG" || value == "DENY" || value == "DENY_LOG"

								if !valid {
									errors = append(errors, fmt.Errorf("invalid outcome value : '%v'", v))
								}

								return
							},
						},
					},
				},
				ConflictsWith: []string{"path"},
			},

			"default_result": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "ALLOWED" || value == "DENIED" || value == "DEFER"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid default result value : '%v'", v))
					}

					return
				},
				ConflictsWith: []string{"path"},
			},

			// AclFile
			"path": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"default_result", "rule"},
			},
		},
	}
}

func createAccessControlProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	attributes := toAccessControlProviderAttributes(d)
	resp, err := client.CreateAccessControlProvider(attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetAccessControlProvider(name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid access control provider'%s': %s", name, resp.Status)
}

func toAccessControlProviderAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceAccessControlProvider().Schema, "rule")

	value, exists := d.GetOk("rule")
	if exists {
		var val, expected = value.([]interface{})
		if expected && val != nil {
			var items = make([]map[string]interface{}, len(val))
			for i, v := range val {
				p := v.(map[string]interface{})
				items[i] = *createMapWithKeysInCameCase(&p)
			}
			(*attributes)["rules"] = items
		}
	} else {
		oldValue, newValue := d.GetChange("rule")
		if fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", newValue) {
			(*attributes)["rules"] = nil
		}
	}
	return attributes
}

func readAccessControlProvider(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetAccessControlProvider(name)
	if err != nil {
		return err
	}

	if len(*attributes) == 0 {
		return nil
	}

	err = applyResourceAttributes(d, attributes, "rule")
	if err != nil {
		return err
	}

	_, keySet := d.GetOk("rule")
	value, attributeSet := (*attributes)[("rules")]
	if keySet || attributeSet {
		if value != nil {
			rules, validType := value.([]interface{})

			if validType {
				val := make([]map[string]interface{}, len(rules))
				for idx, v := range rules {
					p := v.(map[string]interface{})
					val[idx] = (*createMapWithKeysUnderscored(&p))
				}
				value = val
			} else {
				value = nil
			}
		}

		err = d.Set("rule", value)
		if err != nil {
			return err
		}
	}
	return nil
}

func existsAccessControlProvider(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	name := d.Get("name").(string)
	attributes, err := client.GetAccessControlProvider(name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}
	return true, nil
}

func deleteAccessControlProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	resp, err := client.DeleteAccessControlProvider(name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid access control provider '%s': %s", name, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateAccessControlProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toAccessControlProviderAttributes(d)

	resp, err := client.UpdateAccessControlProvider(name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid  access control provider '%s' does not exist", name)
	}

	return fmt.Errorf("error updating qpid access control provider '%s': %s", name, resp.Status)
}
