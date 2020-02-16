package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceVirtualHostNode(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceVirtualHostNodeCheckDestroy(testAcceptanceVirtualHostNodeName),
		Steps: []resource.TestStep{
			{
				// test new virtual host node creation from configuration
				Config: testAcceptanceVirtualHostNodeConfigMinimal,
				Check: testAcceptanceVirtualHostNodeCheck(
					testAcceptanceVirtualHostNodeResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostNodeName, "type": "JSON"},
				),
			},
			{
				// test virtual host node restoration from configuration after its deletion on broker side
				PreConfig: dropVirtualHostNode(testAcceptanceVirtualHostNodeName),
				Config:    testAcceptanceVirtualHostNodeConfigMinimal,
				Check: testAcceptanceVirtualHostNodeCheck(
					testAcceptanceVirtualHostNodeResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostNodeName, "type": "JSON"},
				),
			},
			{
				// test virtual host node update
				Config: getVirtualHostNodeConfigurationWithAttributes(&map[string]string{"context": "{\"foo\":\"bar\"}"}),
				Check: testAcceptanceVirtualHostNodeCheck(
					testAcceptanceVirtualHostNodeResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostNodeName, "type": "JSON", "context": map[string]interface{}{"foo": "bar"}},
				),
			},
			{
				// test virtual host node attribute removal
				Config: testAcceptanceVirtualHostNodeConfigMinimal,
				Check: testAcceptanceVirtualHostNodeCheck(
					testAcceptanceVirtualHostNodeResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostNodeName, "type": "JSON"},
					"context",
				),
			},
		},
	})
}

func dropVirtualHostNode(nodeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteVirtualHostNode(nodeName)
		if err != nil {
			fmt.Printf("unable to delete virtual host node: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete virtual host node: %v", resp))
		}
	}
}

func testAcceptanceVirtualHostNodeCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("virtual host node id not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)
		nodes, err := client.GetVirtualHostNodes()
		if err != nil {
			return fmt.Errorf("error getting nodes: %s", err)
		}

		for _, node := range *nodes {
			if node["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&node, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("virtual host node '%s' is not found", rn)
	}
}

func testAcceptanceVirtualHostNodeCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		nodes, err := client.GetVirtualHostNodes()
		if err != nil {
			return fmt.Errorf("error getting nodes: %s", err)
		}

		for _, node := range *nodes {
			if node["name"] == name {
				return fmt.Errorf("virtual host node '%v' still exists", node)
			}
		}

		return nil
	}
}

const testAcceptanceVirtualHostNodeResourceName = "qpid_virtual_host_node"
const testAcceptanceVirtualHostNodeName = "acceptance_test"
const testAcceptanceVirtualHostNodeResource = testAcceptanceVirtualHostNodeResourceName + "." + testAcceptanceVirtualHostNodeName

const testAcceptanceVirtualHostNodeConfigMinimal = `
resource "` + testAcceptanceVirtualHostNodeResourceName + `" "` + testAcceptanceVirtualHostNodeName + `" {
    name = "` + testAcceptanceVirtualHostNodeName + `"
    type = "JSON"
}`

func getVirtualHostNodeConfigurationWithAttributes(entries *map[string]string) string {
	config := `resource "` + testAcceptanceVirtualHostNodeResourceName + `" "` + testAcceptanceVirtualHostNodeName + `" {
		name = "` + testAcceptanceVirtualHostNodeName + `"
		type = "JSON"
	`
	for k, v := range *entries {
		config += fmt.Sprintf("    %s=%s\n", k, v)
	}
	config += `}
`
	return config
}
