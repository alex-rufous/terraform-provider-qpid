package qpid

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAcceptanceExchange(t *testing.T) {
	var virtualHostNodeName string
	var virtualHostName string
	var exchangeName string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceExchangeCheckDestroy(virtualHostNodeName, virtualHostName, exchangeName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceExchangeConfigMinimal,
				Check: testAcceptanceExchangeCheck(
					"qpid_exchange.test_exchange", &virtualHostNodeName, &virtualHostName, &exchangeName,
				),
			},
			{
				PreConfig: dropExchange("acceptance_test", "acceptance_test_host", "test_exchange"),
				Config:    testAcceptanceExchangeConfigMinimal,
				Check: testAcceptanceExchangeCheck(
					"qpid_exchange.test_exchange", &virtualHostNodeName, &virtualHostName, &exchangeName,
				),
			},
		},
	})
}

func testAcceptanceExchangeCheck(rn string, virtualHostNodeName *string, virtualHostName *string, exchangeName *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("exchange id not set")
		}

		nodeName, ok := rs.Primary.Attributes["parents.0"]
		if !ok {
			return fmt.Errorf("parents: node name is not set")
		}

		hostName, ok := rs.Primary.Attributes["parents.1"]
		if !ok {
			return fmt.Errorf("parents: host name is not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		exchanges, err := client.getVirtualHostExchanges(nodeName, hostName)
		if err != nil {
			return fmt.Errorf("error on getting exchanges: %s", err)
		}

		for _, exchange := range *exchanges {
			if exchange["id"] == rs.Primary.ID {
				*exchangeName = exchange["name"].(string)
				*virtualHostNodeName = nodeName
				*virtualHostName = hostName
				return nil
			}
		}

		return fmt.Errorf("unable to find exchange %s", rn)
	}
}

func testAcceptanceExchangeCheckDestroy(virtualHostNodeName string, virtualHostName string, exchangeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		exchanges, err := client.getVirtualHostExchanges(virtualHostNodeName, virtualHostName)
		if err != nil {
			return fmt.Errorf("error on getting exchanges: %s", err)
		}

		for _, exchange := range *exchanges {
			if exchange["name"] == exchangeName {
				return fmt.Errorf("exchange %s/%s/%s still exist", virtualHostNodeName, virtualHostName, exchangeName)
			}
		}

		return nil
	}
}

func dropExchange(nodeName string, hostName string, exchangeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteExchange(nodeName, hostName, exchangeName)
		if err != nil {
			fmt.Printf("unable to delete exchange : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete exchange: %v", resp))
		}
	}
}

const testAcceptanceExchangeConfigMinimal = `
resource "qpid_virtual_host_node" "acceptance_test" {
    name = "acceptance_test"
    type = "JSON"
    virtual_host_initial_configuration = "{}"
}

resource "qpid_virtual_host" "acceptance_test_host" {
    depends_on = [qpid_virtual_host_node.acceptance_test]
    name = "acceptance_test_host"
    parent = "acceptance_test"
    type = "BDB"
}

resource "qpid_exchange" "test_exchange" {
    name = "test_exchange"
    depends_on = [qpid_virtual_host.acceptance_test_host]
    parents = ["acceptance_test", "acceptance_test_host"]
    type = "direct"
}
`
