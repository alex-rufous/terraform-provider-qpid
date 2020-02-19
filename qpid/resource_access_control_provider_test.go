package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"os"
	"testing"
)

func TestAcceptanceAccessControlProvider(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAcceptancePreCheck(t) },
		Providers: testAcceptanceProviders,
		//CheckDestroy: testAcceptanceAccessControlProviderCheckDestroy(testAcceptanceAccessControlProviderName),
		Steps: []resource.TestStep{
			{
				// test new access control provider creation from configuration
				Config: getRuleBasedAccessControlProviderConfiguration(),
				Check: testAcceptanceAccessControlProviderCheck(
					testAcceptanceAccessControlProviderResource,
					&map[string]interface{}{"name": testAcceptanceAccessControlProviderName, "type": "RuleBased"},
				),
			},
			{
				// test access control provider restoration from configuration after its deletion on broker side
				PreConfig: dropAccessControlProvider(testAcceptanceAccessControlProviderName),
				Config:    getRuleBasedAccessControlProviderConfiguration(),
				Check: testAcceptanceAccessControlProviderCheck(
					testAcceptanceAccessControlProviderResource,
					&map[string]interface{}{"name": testAcceptanceAccessControlProviderName, "type": "RuleBased"},
				),
			},
			{
				// test access control provider update
				Config: getAccessControlProviderConfigurationWithAttributes(&map[string]string{"context": "{\"foo\":\"bar\"}"}),
				Check: testAcceptanceAccessControlProviderCheck(
					testAcceptanceAccessControlProviderResource,
					&map[string]interface{}{"name": testAcceptanceAccessControlProviderName,
						"type":    "RuleBased",
						"context": map[string]interface{}{"foo": "bar"}},
				),
			},
			{
				// test access control provider attribute removal
				Config: getRuleBasedAccessControlProviderConfiguration(),
				Check: testAcceptanceAccessControlProviderCheck(
					testAcceptanceAccessControlProviderResource,
					&map[string]interface{}{"name": testAcceptanceAccessControlProviderName, "type": "RuleBased"},
					"context",
				),
			},
		},
	})
}

func dropAccessControlProvider(nodeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteAccessControlProvider(nodeName)
		if err != nil {
			fmt.Printf("unable to delete access control provider: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete access control provider: %v", resp))
		}
	}
}

func testAcceptanceAccessControlProviderCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("access control provider id not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)
		providers, err := client.GetAccessControlProviders()
		if err != nil {
			return fmt.Errorf("error getting access control provider: %s", err)
		}

		for _, provider := range *providers {
			if provider["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&provider, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("access control provider '%s' is not found", rn)
	}
}

func testAcceptanceAccessControlProviderCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		providers, err := client.GetAccessControlProviders()
		if err != nil {
			return fmt.Errorf("error getting providers: %s", err)
		}

		for _, node := range *providers {
			if node["name"] == name {
				return fmt.Errorf("access control provider '%v' still exists", node)
			}
		}

		return nil
	}
}

const testAcceptanceAccessControlProviderResourceName = "qpid_access_control_provider"
const testAcceptanceAccessControlProviderName = "acceptance_test_access_control_provider"
const testAcceptanceAccessControlProviderName2 = "acceptance_test_access_control_provider2"
const testAcceptanceAccessControlProviderResource = testAcceptanceAccessControlProviderResourceName + "." + testAcceptanceAccessControlProviderName

func getRuleBasedAccessControlProviderConfiguration() string {
	return getAccessControlProviderConfigurationWithAttributes(&map[string]string{})
}

func getAccessControlProviderConfigurationWithAttributes(entries *map[string]string) string {
	config := `
resource "` + testAcceptanceAccessControlProviderResourceName + `" "` + testAcceptanceAccessControlProviderName + `" {
    name = "` + testAcceptanceAccessControlProviderName + `"
    type = "RuleBased"
    priority=1

`
	for k, v := range *entries {
		config += fmt.Sprintf("    %s=%s\n", k, v)
	}

	config += `

    rule {
		identity = "` + os.Getenv("QPID_USERNAME") + `"
        object_type = "ALL"
        operation = "ALL"
        outcome = "ALLOW_LOG"
    }

    rule {
		identity = "` + testAcceptanceGroupResourceName + `"
        object_type = "MANAGEMENT"
        operation = "ACCESS"
        outcome = "ALLOW_LOG"
    }

    rule {
		identity = "` + testAcceptanceGroupResourceName + `"
        object_type = "BROKER"
        operation = "CONFIGURE"
        outcome = "ALLOW_LOG"
    }

    rule {
		identity = "` + testAcceptanceGroupResourceName + `"
        object_type = "ALL"
        operation = "CREATE"
        outcome = "ALLOW_LOG"
    }

    rule {
		identity = "` + testAcceptanceGroupResourceName + `"
        object_type = "ALL"
        operation = "DELETE"
        outcome = "ALLOW_LOG"
    }

    rule {
		identity = "` + testAcceptanceGroupResourceName + `"
        object_type = "ALL"
        operation = "UPDATE"
        outcome = "ALLOW_LOG"
    }

    rule {
		identity = "` + testAcceptanceGroupResourceName + `"
        object_type = "ALL"
        operation = "INVOKE"
        outcome = "ALLOW_LOG"
    }

    rule {
		identity = "ALL"
        object_type = "ALL"
        operation = "ALL"
        outcome = "DENY_LOG"
    }
}

resource "` + testAcceptanceAccessControlProviderResourceName + `" "` + testAcceptanceAccessControlProviderName2 + `" {
    name = "` + testAcceptanceAccessControlProviderName2 + `"
    type = "AllowAll"
    priority=10
}

`

	return config
}
