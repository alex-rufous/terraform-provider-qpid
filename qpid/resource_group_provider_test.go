package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceGroupProvider(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceGroupProviderCheckDestroy(testAcceptanceGroupProviderName),
		Steps: []resource.TestStep{
			{
				// test new group provider creation from configuration
				Config: testAcceptanceGroupProviderManagedGroupProviderConfigMinimal,
				Check: testAcceptanceGroupProviderCheck(
					testAcceptanceGroupProviderResource,
					&map[string]interface{}{"name": testAcceptanceGroupProviderName, "type": "ManagedGroupProvider"},
				),
			},
			{
				// test group provider restoration from configuration after its deletion on broker side
				PreConfig: dropGroupProvider(testAcceptanceGroupProviderName),
				Config:    testAcceptanceGroupProviderManagedGroupProviderConfigMinimal,
				Check: testAcceptanceGroupProviderCheck(
					testAcceptanceGroupProviderResource,
					&map[string]interface{}{"name": testAcceptanceGroupProviderName, "type": "ManagedGroupProvider"},
				),
			},
			{
				// test group provider update with new attribute
				Config: getGroupProviderConfigurationWithAttributes(&map[string]string{"context": "{\"foo\":\"bar\"}"}),
				Check: testAcceptanceGroupProviderCheck(
					testAcceptanceGroupProviderResource,
					&map[string]interface{}{"name": testAcceptanceGroupProviderName,
						"type":    "ManagedGroupProvider",
						"context": map[string]interface{}{"foo": "bar"}},
				),
			},
			{
				// test group provider attribute removal
				Config: testAcceptanceGroupProviderManagedGroupProviderConfigMinimal,
				Check: testAcceptanceGroupProviderCheck(
					testAcceptanceGroupProviderResource,
					&map[string]interface{}{"name": testAcceptanceGroupProviderName, "type": "ManagedGroupProvider"},
					"context",
				),
			},
		},
	})
}

func dropGroupProvider(nodeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteGroupProvider(nodeName)
		if err != nil {
			fmt.Printf("unable to delete group provider: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete group provider: %v", resp))
		}
	}
}

func testAcceptanceGroupProviderCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("group provider id not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)
		nodes, err := client.GetGroupProviders()
		if err != nil {
			return fmt.Errorf("error getting group provider: %s", err)
		}

		for _, node := range *nodes {
			if node["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&node, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("group provider '%s' is not found", rn)
	}
}

func testAcceptanceGroupProviderCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		nodes, err := client.GetGroupProviders()
		if err != nil {
			return fmt.Errorf("error getting nodes: %s", err)
		}

		for _, node := range *nodes {
			if node["name"] == name {
				return fmt.Errorf("group provider '%v' still exists", node)
			}
		}

		return nil
	}
}

const testAcceptanceGroupProviderResourceName = "qpid_group_provider"
const testAcceptanceGroupProviderName = "test_group_provider"
const testAcceptanceGroupProviderResource = testAcceptanceGroupProviderResourceName + "." + testAcceptanceGroupProviderName

const testAcceptanceGroupProviderManagedGroupProviderConfigMinimal = `
resource "` + testAcceptanceGroupProviderResourceName + `" "` + testAcceptanceGroupProviderName + `" {
    name = "` + testAcceptanceGroupProviderName + `"
    type = "ManagedGroupProvider"
}
`

func getGroupProviderConfigurationWithAttributes(entries *map[string]string) string {
	config := `
resource "` + testAcceptanceGroupProviderResourceName + `" "` + testAcceptanceGroupProviderName + `" {
    name = "` + testAcceptanceGroupProviderName + `"
    type = "ManagedGroupProvider"
`
	for k, v := range *entries {
		config += fmt.Sprintf("    %s=%s\n", k, v)
	}
	config += `}
`

	return config
}
