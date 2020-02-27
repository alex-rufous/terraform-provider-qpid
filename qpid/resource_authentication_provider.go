package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceAuthenticationProvider() *schema.Resource {

	return &schema.Resource{
		Create: createAuthenticationProvider,
		Read:   readAuthenticationProvider,
		Delete: deleteAuthenticationProvider,
		Update: updateAuthenticationProvider,
		Exists: existsAuthenticationProvider,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of authentication provider",
				Required:    true,
				ForceNew:    true,
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of authentication provider",
				Required:    true,
				ForceNew:    true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "SimpleLDAP" || value == "PlainPasswordFile" ||
						value == "Base64MD5PasswordFile" || value == "SCRAM-SHA-256" || value == "SCRAM-SHA-1" ||
						value == "External" || value == "OAuth2" || value == "Kerberos" || value == "Anonymous" ||
						value == "Plain" || value == "MD5"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid authentication provider type value : '%v'", v))
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

			"secure_only_mechanisms": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},

			"disabled_mechanisms": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},

			// External provider fields
			"use_full_d_n": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"path", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "trust_store", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			// Base64MD5PasswordFile && PlainPasswordFile  provider fields
			"path": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "trust_store", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			// SimpleLDAP provider fields
			"provider_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"provider_auth_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"search_context": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"search_filter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"bind_without_search": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"ldap_context_factory": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"trust_store": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"search_username": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"search_password": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"group_attribute_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"group_search_context": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"group_search_filter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"group_subtree_search_scope": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"authentication_method": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					valid := value == "NONE" || value == "SIMPLE" || value == "GSSAPI"

					if !valid {
						errors = append(errors, fmt.Errorf("invalid authentication method value : '%v'. Allowed values: \"NONE\", \"SIMPLE\", \"GSSAPI\"", v))
					}

					return
				},
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			"login_config_scope": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "path", "authorization_endpoint_u_r_i",
					"token_endpoint_u_r_i", "token_endpoint_needs_auth", "identity_resolver_endpoint_u_r_i",
					"identity_resolver_type", "post_logout_u_r_i", "client_id", "client_secret", "scope"},
			},

			// OAuth2 provider fields
			"authorization_endpoint_u_r_i": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"token_endpoint_u_r_i": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"token_endpoint_needs_auth": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"identity_resolver_endpoint_u_r_i": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"identity_resolver_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"post_logout_u_r_i": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"client_id": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"client_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				ForceNew:  false,
				Default:   nil,
				Sensitive: true,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},

			"scope": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
				Default:  nil,
				ConflictsWith: []string{"use_full_d_n", "provider_url", "provider_auth_url", "search_context", "search_filter",
					"bind_without_search", "ldap_context_factory", "search_username", "search_password",
					"group_attribute_name", "group_search_context", "group_search_filter", "group_subtree_search_scope",
					"authentication_method", "login_config_scope"},
			},
		},
	}
}

func createAuthenticationProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)
	attributes := toAuthenticationProviderAttributes(d)
	resp, err := client.CreateAuthenticationProvider(attributes)
	if err != nil {
		return err
	}

	name := (*attributes)["name"].(string)
	if resp.StatusCode == http.StatusCreated {
		attributes, err := convertHttpResponseToMap(resp)
		if err != nil {
			var err2 error
			attributes, err2 = client.GetAuthenticationProvider(name)
			if err2 != nil {
				return err
			}
		}
		id := (*attributes)["id"].(string)
		d.SetId(id)
		return nil
	}

	return fmt.Errorf("error creating qpid authentication provider'%s': %s", name, resp.Status)
}

func toAuthenticationProviderAttributes(d *schema.ResourceData) *map[string]interface{} {
	attributes := schemaToAttributes(d, resourceAuthenticationProvider().Schema)
	return attributes
}

func readAuthenticationProvider(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*Client)

	name := d.Get("name").(string)
	attributes, err := client.GetAuthenticationProvider(name)
	if err != nil {
		return err
	}

	return applyResourceAttributes(d, attributes)
}

func existsAuthenticationProvider(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*Client)
	name := d.Get("name").(string)
	attributes, err := client.GetAuthenticationProvider(name)
	if err != nil {
		return false, err
	}
	if len(*attributes) == 0 {
		return false, nil
	}
	return true, nil
}

func deleteAuthenticationProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)
	resp, err := client.DeleteAuthenticationProvider(name)
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("error deleting qpid authentication provider '%s': %s", name, resp.Status)
	}
	d.SetId("")
	return nil
}

func updateAuthenticationProvider(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client)

	name := d.Get("name").(string)

	attributes := toAuthenticationProviderAttributes(d)

	resp, err := client.UpdateAuthenticationProvider(name, attributes)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("qpid  authentication provider '%s' does not exist", name)
	}

	return fmt.Errorf("error updating qpid authentication provider '%s': %s", name, resp.Status)
}
