package qpid

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Running instance Qpid broker is required to run the acceptance tests
//
// The environment variables QPID_ENDPOINT, QPID_USERNAME, QPID_PASSWORD and QPID_MODEL_VERSION
// needs to be set in order to run the tests.
//
// Optionally, QPID_CERTIFICATE and QPID_SKIP_CERT_VERIFICATION can be set for a self-signed certificate
//
// The tests can be run like below
//    make testacc
//
// Alternatively, the acceptance tests can be executed as below
//    go test $(go list ./... |grep -v 'vendor') -v  -timeout 120m

var testAcceptanceProviders map[string]terraform.ResourceProvider
var testAcceptanceProvider *schema.Provider

func init() {
	testAcceptanceProvider = Provider().(*schema.Provider)
	testAcceptanceProviders = map[string]terraform.ResourceProvider{
		"qpid": testAcceptanceProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("error: %s", err)
	}
}

func testAcceptancePreCheck(t *testing.T) {
	for _, name := range []string{"QPID_ENDPOINT", "QPID_USERNAME", "QPID_PASSWORD", "QPID_MODEL_VERSION"} {
		if v := os.Getenv(name); v == "" {
			t.Fatal("QPID_ENDPOINT, QPID_USERNAME, QPID_PASSWORD and QPID_MODEL_VERSION must be set for acceptance tests")
		}
	}
}
