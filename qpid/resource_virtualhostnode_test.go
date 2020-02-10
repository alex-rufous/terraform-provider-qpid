package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceVirtualHostNode(t *testing.T) {
	var virtualHostNodeName string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceVirtualHostNodeCheckDestroy(virtualHostNodeName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceVirtualHostNodeConfigMinimal,
				Check: testAcceptanceVirtualHostNodeCheck(
					"qpid_virtual_host_node.test", &virtualHostNodeName,
				),
			},
			{
				PreConfig: dropVirtualHostNode(&virtualHostNodeName),
				Config:    testAcceptanceVirtualHostNodeConfigMinimal,
				Check: testAcceptanceVirtualHostNodeCheck(
					"qpid_virtual_host_node.test", &virtualHostNodeName,
				),
			},
		},
	})
}

func dropVirtualHostNode(nodeName *string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteVirtualHostNode(*nodeName)
		if err != nil {
			fmt.Printf("unable to delete virtual host node: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete virtual host node: %v", resp))
		}
	}
}

func testAcceptanceVirtualHostNodeCheck(rn string, name *string) resource.TestCheckFunc {
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
				*name = node["name"].(string)
				return nil
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

const testAcceptanceVirtualHostNodeConfigMinimal = `
resource "qpid_virtual_host_node" "test" {
    name = "test"
    type = "JSON"
    virtual_host_initial_configuration = "{\"type\":\"BDB\"}"
}`
