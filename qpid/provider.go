package qpid

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"io/ioutil"
	"net/http"
)

// Provider ...
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("QPID_ENDPOINT", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("endpoint must not be an empty string"))
					}

					return
				},
			},

			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("QPID_USERNAME", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("username must not be an empty string"))
					}

					return
				},
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("QPID_PASSWORD", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("password must not be an empty string"))
					}

					return
				},
			},

			"model_version": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("QPID_MODEL_VERSION", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("model version must not be an empty string"))
					}

					return
				},
			},

			"skip_cert_verification": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("QPID_SKIP_CERT_VERIFICATION", nil),
			},

			"certificate_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("QPID_CERTIFICATE", ""),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"qpid_virtual_host_node":       resourceVirtualHostNode(),
			"qpid_virtual_host":            resourceVirtualHost(),
			"qpid_queue":                   resourceQueue(),
			"qpid_exchange":                resourceExchange(),
			"qpid_binding":                 resourceBinding(),
			"qpid_authentication_provider": resourceAuthenticationProvider(),
			"qpid_user":                    resourceUser(),
			"qpid_group_provider":          resourceGroupProvider(),
			"qpid_group":                   resourceGroup(),
			"qpid_group_member":            resourceGroupMember(),
			"qpid_access_control_provider": resourceAccessControlProvider(),
			"qpid_key_store":               resourceKeyStore(),
			"qpid_trust_store":             resourceTrustStore(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var username = d.Get("username").(string)
	var password = d.Get("password").(string)
	var endpoint = d.Get("endpoint").(string)
	var modelVersion = d.Get("model_version").(string)
	var skipCertVerification = d.Get("skip_cert_verification").(bool)
	var certificateFile = d.Get("certificate_file").(string)

	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = skipCertVerification
	if certificateFile != "" {
		cert, err := ioutil.ReadFile(certificateFile)
		if err != nil {
			return nil, err
		}

		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cert)
		tlsConfig.RootCAs = certPool
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client, err := NewClient(endpoint, username, password, modelVersion, transport)
	if err != nil {
		return nil, err
	}

	return client, nil
}
