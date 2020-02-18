package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceAuthenticationProvider(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceAuthenticationProviderCheckDestroy(testAcceptanceAuthenticationProviderName),
		Steps: []resource.TestStep{
			{
				// test new authentication provider creation from configuration
				Config: testAcceptanceAuthenticationProviderPlainConfigMinimal,
				Check: testAcceptanceAuthenticationProviderCheck(
					testAcceptanceAuthenticationProviderResource,
					&map[string]interface{}{"name": testAcceptanceAuthenticationProviderName, "type": "Plain"},
				),
			},
			{
				// test authentication provider restoration from configuration after its deletion on broker side
				PreConfig: dropAuthenticationProvider(testAcceptanceAuthenticationProviderName),
				Config:    testAcceptanceAuthenticationProviderPlainConfigMinimal,
				Check: testAcceptanceAuthenticationProviderCheck(
					testAcceptanceAuthenticationProviderResource,
					&map[string]interface{}{"name": testAcceptanceAuthenticationProviderName, "type": "Plain"},
				),
			},
			{
				// test authentication provider update
				Config: getAuthenticationProviderConfigurationWithAttributes(&map[string]string{"secure_only_mechanisms": "[\"PLAIN\", \"CRAM-MD5\", \"SCRAM-SHA-1\"]"}),
				Check: testAcceptanceAuthenticationProviderCheck(
					testAcceptanceAuthenticationProviderResource,
					&map[string]interface{}{"name": testAcceptanceAuthenticationProviderName,
						"type":                 "Plain",
						"secureOnlyMechanisms": []interface{}{"PLAIN", "CRAM-MD5", "SCRAM-SHA-1"}},
				),
			},
			{
				// test authentication provider attribute removal
				Config: testAcceptanceAuthenticationProviderPlainConfigMinimal,
				Check: testAcceptanceAuthenticationProviderCheck(
					testAcceptanceAuthenticationProviderResource,
					&map[string]interface{}{"name": testAcceptanceAuthenticationProviderName, "type": "Plain"},
					"context",
				),
			},
		},
	})
}

func dropAuthenticationProvider(nodeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteAuthenticationProvider(nodeName)
		if err != nil {
			fmt.Printf("unable to delete authentication provider: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete authentication provider: %v", resp))
		}
	}
}

func testAcceptanceAuthenticationProviderCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("authentication provider id not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)
		nodes, err := client.GetAuthenticationProviders()
		if err != nil {
			return fmt.Errorf("error getting authentication provider: %s", err)
		}

		for _, node := range *nodes {
			if node["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&node, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("authentication provider '%s' is not found", rn)
	}
}

func testAcceptanceAuthenticationProviderCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		nodes, err := client.GetAuthenticationProviders()
		if err != nil {
			return fmt.Errorf("error getting nodes: %s", err)
		}

		for _, node := range *nodes {
			if node["name"] == name {
				return fmt.Errorf("authentication provider '%v' still exists", node)
			}
		}

		return nil
	}
}

const testAcceptanceAuthenticationProviderResourceName = "qpid_authentication_provider"
const testAcceptanceAuthenticationProviderName = "test_authentication_provider"
const testAcceptanceAuthenticationProviderResource = testAcceptanceAuthenticationProviderResourceName + "." + testAcceptanceAuthenticationProviderName

const testAcceptanceAuthenticationProviderPlainConfigMinimal = `
resource "` + testAcceptanceAuthenticationProviderResourceName + `" "` + testAcceptanceAuthenticationProviderName + `" {
    name = "` + testAcceptanceAuthenticationProviderName + `"
    type = "Plain"
}
`

func getAuthenticationProviderConfigurationWithAttributes(entries *map[string]string) string {
	config := `
resource "` + testAcceptanceAuthenticationProviderResourceName + `" "` + testAcceptanceAuthenticationProviderName + `" {
    name = "` + testAcceptanceAuthenticationProviderName + `"
    type = "Plain"
`
	for k, v := range *entries {
		config += fmt.Sprintf("    %s=%s\n", k, v)
	}
	config += `}
`

	return config
}
