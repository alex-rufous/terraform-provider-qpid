package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceVirtualHost(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceVirtualHostCheckDestroy(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceVirtualHostConfigMinimal,
				Check: testAcceptanceVirtualHostCheck(
					testAcceptanceVirtualHostResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostName, "type": "BDB"},
				),
			},
			{
				PreConfig: dropVirtualHost(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName),
				Config:    testAcceptanceVirtualHostConfigMinimal,
				Check: testAcceptanceVirtualHostCheck(
					testAcceptanceVirtualHostResource,
					&map[string]interface{}{"name": testAcceptanceVirtualHostName, "type": "BDB"},
				),
			},

			{
				Config: testAcceptanceVirtualHostParent + `

resource "` + testAcceptanceVirtualHostResourceName + `" "` + testAcceptanceVirtualHostName + `" {
    depends_on = [` + testAcceptanceVirtualHostNodeResource + `]
    name = "` + testAcceptanceVirtualHostName + `"
    virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    type = "BDB"

    node_auto_creation_policy {
                                   pattern = ".*"
                                   created_on_publish = true
                                   created_on_consume = true
                                   node_type = "Queue"
                                   attributes = {}
                              }
}`,
				Check: testAcceptanceVirtualHostCheck(
					testAcceptanceVirtualHostResource,
					&map[string]interface{}{"nodeAutoCreationPolicies": []interface{}{
						map[string]interface{}{
							"pattern":          ".*",
							"createdOnPublish": true,
							"createdOnConsume": true,
							"nodeType":         "Queue",
							"attributes":       map[string]interface{}{}}}},
				),
			},

			{
				Config: getVirtualHostConfigurationWithAttributes(&map[string]string{"store_overfull_size": "100000"}),
				Check: testAcceptanceVirtualHostCheck(
					testAcceptanceVirtualHostResource,
					&map[string]interface{}{"storeOverfullSize": 100000.0},
					"nodeAutoCreationPolicies",
				),
			},
		},
	})
}

func testAcceptanceVirtualHostCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("virtual host id not set")
		}

		nodeName, ok := rs.Primary.Attributes["virtual_host_node"]
		if !ok {
			return fmt.Errorf("virtual_host_node not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		hosts, err := client.GetNodeVirtualHosts(nodeName)
		if err != nil {
			return fmt.Errorf("error on getting hosts: %s", err)
		}

		for _, host := range *hosts {
			if host["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&host, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find virtualhost %s", rn)
	}
}

func testAcceptanceVirtualHostCheckDestroy(virtualHostNodeName string, virtualHostName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		hosts, err := client.GetNodeVirtualHosts(virtualHostNodeName)
		if err != nil {
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

const testAcceptanceVirtualHostName = "acceptance_test_host"
const testAcceptanceVirtualHostResourceName = "qpid_virtual_host"
const testAcceptanceVirtualHostResource = testAcceptanceVirtualHostResourceName + "." + testAcceptanceVirtualHostName

const testAcceptanceVirtualHostParent = `
resource "` + testAcceptanceVirtualHostNodeResourceName + `" "` + testAcceptanceVirtualHostNodeName + `" {
    name = "` + testAcceptanceVirtualHostNodeName + `"
    type = "JSON"
    virtual_host_initial_configuration = "{}"
}
`
const testAcceptanceVirtualHostConfigMinimal = testAcceptanceVirtualHostParent + `
resource "` + testAcceptanceVirtualHostResourceName + `" "` + testAcceptanceVirtualHostName + `" {
    depends_on = [` + testAcceptanceVirtualHostNodeResource + `]
    name = "` + testAcceptanceVirtualHostName + `"
    virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    type = "BDB"
}`

func getVirtualHostConfigurationWithAttributes(entries *map[string]string) string {
	config := testAcceptanceVirtualHostParent + `
resource "` + testAcceptanceVirtualHostResourceName + `" "` + testAcceptanceVirtualHostName + `" {
    depends_on = [` + testAcceptanceVirtualHostNodeResource + `]
    name = "` + testAcceptanceVirtualHostName + `"
    virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    type = "BDB"
`
	for k, v := range *entries {
		config += fmt.Sprintf("    %s=%s\n", k, v)
	}
	config += `}
`
	return config
}
