package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceVirtualHostAlias(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceVirtualHostAliasCheckDestroy(testAcceptanceAmqpPortConfigMinimal, testAcceptanceVirtualHostAliasName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceVirtualHostAliasConfigMinimal,
				Check: testAcceptanceVirtualHostAliasCheck(
					testAcceptanceVirtualHostAliasResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostAliasName, "type": "patternMatchingAlias"},
				),
			},
			{
				PreConfig: dropVirtualHostAlias(testAcceptancePortName, testAcceptanceVirtualHostAliasName),
				Config:    testAcceptanceVirtualHostAliasConfigMinimal,
				Check: testAcceptanceVirtualHostAliasCheck(
					testAcceptanceVirtualHostAliasResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostAliasName, "type": "patternMatchingAlias"},
				),
			},

			{
				Config: testAcceptanceAmqpPortConfigMinimal + `

resource "` + testAcceptanceVirtualHostAliasResourceName + `" "` + testAcceptanceVirtualHostAliasName + `" {
    depends_on = [` + testAcceptancePortResource + `]
    name = "` + testAcceptanceVirtualHostAliasName + `"
    port = "` + testAcceptancePortName + `"
    type = "patternMatchingAlias"
    priority = 9999
	pattern = ".*"
}`,
				Check: testAcceptanceVirtualHostAliasCheck(
					testAcceptanceVirtualHostAliasResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostAliasName, "type": "patternMatchingAlias", "priority": 9999.0},
				),
			},
		},
	})
}

func testAcceptanceVirtualHostAliasCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("alias id not set")
		}

		aliasProvider, ok := rs.Primary.Attributes["port"]
		if !ok {
			return fmt.Errorf("port not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		aliases, err := client.GetVirtualHostAliases(aliasProvider)

		if err != nil {
			return fmt.Errorf("error on getting aliass: %s", err)
		}

		for _, alias := range *aliases {
			if alias["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&alias, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find alias %s", rn)
	}
}

func testAcceptanceVirtualHostAliasCheckDestroy(testVirtualHostAliasProviderName string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		aliases, err := client.GetVirtualHostAliases(testVirtualHostAliasProviderName)
		if err != nil {
			return fmt.Errorf("error on getting aliass for alias provider '%s' : %s", testVirtualHostAliasProviderName, err)
		}

		for _, alias := range *aliases {
			if alias["name"] == name {
				return fmt.Errorf("alias %s/%s still exist", testVirtualHostAliasProviderName, name)
			}
		}

		return nil
	}
}

func dropVirtualHostAlias(portName string, aliasName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteVirtualHostAlias(portName, aliasName)
		if err != nil {
			fmt.Printf("unable to delete alias : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete alias: %v", resp))
		}
	}
}

const testAcceptanceVirtualHostAliasName = "acceptance_test_virtual_host_alias"
const testAcceptanceVirtualHostAliasResourceName = "qpid_virtual_host_alias"
const testAcceptanceVirtualHostAliasResource = testAcceptanceVirtualHostAliasResourceName + "." + testAcceptanceVirtualHostAliasName
const testAcceptanceVirtualHostAliasConfigMinimal = testAcceptanceAmqpPortConfigMinimal + `
resource "` + testAcceptanceVirtualHostAliasResourceName + `" "` + testAcceptanceVirtualHostAliasName + `" {
    depends_on = [` + testAcceptancePortResource + `]
    name = "` + testAcceptanceVirtualHostAliasName + `"
    port = "` + testAcceptancePortName + `"
    type = "patternMatchingAlias"
    pattern = "^` + testAcceptanceVirtualHostName + `$"
}`
