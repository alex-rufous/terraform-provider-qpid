package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"strings"
	"testing"
)

func TestAcceptanceVirtualHost(t *testing.T) {
	var virtualHostNodeName string
	var virtualHostName string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceVirtualHostCheckDestroy(virtualHostNodeName, virtualHostName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceVirtualHostConfigMinimal,
				Check: testAcceptanceVirtualHostCheck(
					"qpid_virtual_host.bar", &virtualHostNodeName, &virtualHostName,
				),
			},
			{
				PreConfig: dropVirtualHost(virtualHostNodeName, virtualHostName),
				Config:    testAcceptanceVirtualHostConfigMinimal,
				Check: testAcceptanceVirtualHostCheck(
					"qpid_virtual_host.bar", &virtualHostNodeName, &virtualHostName,
				),
			},
		},
	})
}

func testAcceptanceVirtualHostCheck(rn string, virtualHostNodeName *string, virtualHostName *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("virtual host id not set")
		}

		nodeName, ok := rs.Primary.Attributes["parent"]
		if !ok {
			return fmt.Errorf("parent not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		hosts, err := client.GetNodeVirtualHosts(nodeName)
		if err != nil {
			return fmt.Errorf("error on getting hosts: %s", err)
		}

		for _, host := range *hosts {
			if host["id"] == rs.Primary.ID {
				*virtualHostName = host["name"].(string)
				*virtualHostNodeName = nodeName
				return nil
			}
		}

		return fmt.Errorf("unable to find virtualhost %s", rn)
	}
}

func testAcceptanceVirtualHostCheckDestroy(virtualHostNodeName string, virtualHostName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		hosts, err := client.GetNodeVirtualHosts(virtualHostNodeName)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return fmt.Errorf("error on getting virtual hosts for node '%s' : %s", virtualHostName, err)
		}

		for _, host := range *hosts {
			if host["name"] == virtualHostName {
				return fmt.Errorf("virtual host %s/%s still exist", virtualHostNodeName, virtualHostName)
			}
		}

		return nil
	}
}

func dropVirtualHost(nodeName string, hostName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteVirtualHost(nodeName, hostName)
		if err != nil {
			fmt.Printf("unable to delete virtual host : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete virtual host node: %v", resp))
		}
	}
}

const testAcceptanceVirtualHostConfigMinimal = `
resource "qpid_virtual_host_node" "foo" {
    name = "foo"
    type = "JSON"
    virtual_host_initial_configuration = "{}"
}

resource "qpid_virtual_host" "bar" {
    depends_on = [qpid_virtual_host_node.foo]
    name = "bar"
    parent = "foo"
    type = "BDB"
    node_auto_creation_policy {
                                   pattern = ".*"
                                   created_on_publish = true
                                   created_on_consume = true
                                   node_type = "Queue"
                                   attributes = {}
                              }
}`
