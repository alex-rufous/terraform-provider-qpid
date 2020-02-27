package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceTrustStore() *schema.Resource {

	return &schema.Resource{
		Create: createTrustStore,
		Read:   readTrustStore,
		Delete: deleteTrustStore,
		Update: updateTrustStore,
		Exists: existsTrustStore,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of trust store",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of trust store",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "FileTrustStore" || value == "NonJavaTrustStore" || value == "SiteSpecificTrustStore"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid trust store type value : '%v'", v))
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
			"trust_anchor_validity_enforced": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"exposed_as_message_source": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
			},

			"included_virtual_host_node_message_sources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"excluded_virtual_host_node_message_sources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// FileTrustStore
			"store_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				Sensitive:     true,
				ConflictsWith: []string{"site_url", "certificates_url"},
			},

			"trust_manager_factory_algorithm": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"site_url", "certificates_url"},
			},
			"trust_store_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"site_url", "certificates_url"},
			},
			"password": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				Sensitive:     true,
				ConflictsWith: []string{"site_url", "certificates_url"},
			},

			"peers_only": {
				Type:          schema.TypeBool,
				Optional:      true,
				ForceNew:      false,
				Default:       nil,
				ConflictsWith: []string{"site_url", "certificates_url"},
			},

			// NonJavaTrustStore
			"certificates_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"site_url", "peers_only", "password", "trust_store_type",
					"trust_manager_factory_algorithm", "store_url"},
			},

			// SiteSpecificTrustStore
			"site_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"certificates_url", "peers_only", "password", "trust_store_type",
					"trust_manager_factory_algorithm", "store_url"},
			},
		},
	}
}

func createTrustStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	attributes := toTrustStoreAttributes(d)

	resp, err := client.CreateTrustStore(attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetTrustStore(name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	m, _ := getErrorResponse(resp)

	return fmt.Errorf("error creating qpid trust store'%s': %s, %v", name, resp.Status, m)
}

func toTrustStoreAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceTrustStore().Schema)
	return attributes
}

func readTrustStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetTrustStore(name)
	if err != nil {
		return err
	}

	return applyResourceAttributes(d, attributes)
}

func existsTrustStore(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	name := d.Get("name").(string)
	attributes, err := client.GetTrustStore(name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}
	return true, nil
}

func deleteTrustStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	resp, err := client.DeleteTrustStore(name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid trust store '%s': %s", name, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateTrustStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toTrustStoreAttributes(d)
	resp, err := client.UpdateTrustStore(name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid  trust store '%s' does not exist", name)
	}

	m, _ := getErrorResponse(resp)

	return fmt.Errorf("error updating qpid trust store '%s': %s : %v", name, resp.Status, m)
}
