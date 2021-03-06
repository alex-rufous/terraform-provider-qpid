package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceKeyStore() *schema.Resource {

	return &schema.Resource{
		Create: createKeyStore,
		Read:   readKeyStore,
		Delete: deleteKeyStore,
		Update: updateKeyStore,
		Exists: existsKeyStore,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of key store",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of key store",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "FileKeyStore" || value == "AutoGeneratedSelfSigned" || value == "NonJavaKeyStore"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid key store type value : '%v'", v))
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

			// FileKeyStore
			"store_url": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				ConflictsWith: []string{"private_key_url", "certificate_url", "intermediate_certificate_url",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},
			"certificate_alias": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"private_key_url", "certificate_url", "intermediate_certificate_url",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},
			"key_manager_factory_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"private_key_url", "certificate_url", "intermediate_certificate_url",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},
			"key_store_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"private_key_url", "certificate_url", "intermediate_certificate_url",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				ConflictsWith: []string{"private_key_url", "certificate_url", "intermediate_certificate_url",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},
			"use_host_name_matching": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"private_key_url", "certificate_url", "intermediate_certificate_url",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},

			// NonJavaKeyStore
			"private_key_url": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				ConflictsWith: []string{"store_url", "certificate_alias", "key_manager_factory_algorithm",
					"key_store_type", "password", "use_host_name_matching",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},

			"certificate_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"store_url", "certificate_alias", "key_manager_factory_algorithm",
					"key_store_type", "password", "use_host_name_matching",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},
			"intermediate_certificate_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"store_url", "certificate_alias", "key_manager_factory_algorithm",
					"key_store_type", "password", "use_host_name_matching",
					"key_algorithm", "signature_algorithm", "key_length", "duration_in_months"},
			},

			// AutoGeneratedSelfSigned
			"key_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"store_url", "certificate_alias", "key_manager_factory_algorithm",
					"key_store_type", "password", "use_host_name_matching",
					"private_key_url", "certificate_url", "intermediate_certificate_url"},
			},

			"signature_algorithm": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"store_url", "certificate_alias", "key_manager_factory_algorithm",
					"key_store_type", "password", "use_host_name_matching",
					"private_key_url", "certificate_url", "intermediate_certificate_url"},
			},

			"key_length": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"store_url", "certificate_alias", "key_manager_factory_algorithm",
					"key_store_type", "password", "use_host_name_matching",
					"private_key_url", "certificate_url", "intermediate_certificate_url"},
			},

			"duration_in_months": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"store_url", "certificate_alias", "key_manager_factory_algorithm",
					"key_store_type", "password", "use_host_name_matching",
					"private_key_url", "certificate_url", "intermediate_certificate_url"},
			},
		},
	}
}

func createKeyStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	attributes := toKeyStoreAttributes(d)

	resp, err := client.CreateKeyStore(attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetKeyStore(name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	m, _ := getErrorResponse(resp)

	return fmt.Errorf("error creating qpid key store'%s': %s, %v", name, resp.Status, m)
}

func toKeyStoreAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceKeyStore().Schema)
	return attributes
}

func readKeyStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetKeyStore(name)
	if err != nil {
		return err
	}

	return applyResourceAttributes(d, attributes)
}

func existsKeyStore(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	name := d.Get("name").(string)
	attributes, err := client.GetKeyStore(name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}
	return true, nil
}

func deleteKeyStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	resp, err := client.DeleteKeyStore(name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid key store '%s': %s", name, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateKeyStore(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toKeyStoreAttributes(d)
	resp, err := client.UpdateKeyStore(name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid  key store '%s' does not exist", name)
	}

	m, _ := getErrorResponse(resp)

	return fmt.Errorf("error updating qpid key store '%s': %s : %v", name, resp.Status, m)
}
