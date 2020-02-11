package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"strings"
	"testing"
)

func TestAcceptanceBinding(t *testing.T) {

	var binding *Binding

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceBindingCheckDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceBindingConfigMinimal,
				Check: testAcceptanceBindingCheck(
					"qpid_binding.test_binding", binding,
				),
			},
			{
				PreConfig: dropBinding(testNodeName, testHostName, testExchangeName, testQueue, testBindingKey),
				Config:    testAcceptanceBindingConfigMinimal,
				Check: testAcceptanceBindingCheck(
					"qpid_binding.test_binding", binding,
				),
			},
		},
	})
}

func testAcceptanceBindingCheck(rn string, binding *Binding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("binding id not set")
		}

		nodeName, ok := rs.Primary.Attributes["parents.0"]
		if !ok {
			return fmt.Errorf("parents: node name is not set")
		}

		hostName, ok := rs.Primary.Attributes["parents.1"]
		if !ok {
			return fmt.Errorf("parents: host name is not set")
		}

		destination, ok := rs.Primary.Attributes["destination"]
		if !ok {
			return fmt.Errorf("destination is not set")
		}

		exchange, ok := rs.Primary.Attributes["exchange"]
		if !ok {
			return fmt.Errorf("exchange is not set")
		}

		id := rs.Primary.ID
		parts := strings.Split(id, "|")
		if len(parts) != 5 {
			return fmt.Errorf("unexpected id '%s' for qpid binding", id)
		}

		exch := parts[2]
		key := parts[4]
		dest := parts[3]
		node := parts[0]
		host := parts[1]

		if dest != destination {
			return fmt.Errorf("unexpected ID: destination does not match")
		}

		if exchange != exch {
			return fmt.Errorf("unexpected ID: exchange does not match")
		}

		if node != nodeName {
			return fmt.Errorf("unexpected ID: node name does not match")
		}

		if host != hostName {
			return fmt.Errorf("unexpected ID: host name does not match")
		}

		if !ok {
			return fmt.Errorf("exchange is not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		bindings, err := client.getExchangeBindings(nodeName, hostName, exchange)
		if err != nil {
			return fmt.Errorf("error on getting queues: %s", err)
		}

		for _, bnd := range *bindings {
			bindingKey := bnd["bindingKey"]
			bindingDestination := bnd["destination"]
			if bindingKey != nil && bindingKey == key && bindingDestination != nil && bindingDestination == dest {
				args := bnd["arguments"]
				var arguments map[string]string
				if args != nil {
					i := args.(map[string]interface{})
					arguments = *convertToMapOfStrings(&i)
				}
				binding = &Binding{bindingKey.(string),
					bindingDestination.(string),
					exchange,
					arguments,
					nodeName,
					hostName}
				return nil
			}
		}

		return fmt.Errorf("unable to find binding %s", rn)
	}
}

func testAcceptanceBindingCheckDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		bindings, err := client.getExchangeBindings(testNodeName, testHostName, testExchangeName)
		if err != nil {
			return fmt.Errorf("error on getting bindings: %s", err)
		}

		for _, bnd := range *bindings {
			if bnd["bindingKey"] == testBindingKey && bnd["destination"] == testQueue {
				return fmt.Errorf("binding %s/%s/%s/%s/%s still exist", testNodeName, testHostName, testExchangeName, testQueue, testBindingKey)
			}
		}

		return nil
	}
}

func dropBinding(nodeName string, hostName string, exchange string, queueName string, bindingKey string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteBinding(&Binding{bindingKey, queueName, exchange, nil, nodeName, hostName})
		if err != nil {
			fmt.Printf("unable to delete binding : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete binding: %v", resp))
		}
	}
}

const testNodeName = "acceptance_test"
const testHostName = "acceptance_test_host"
const testExchangeName = "test_exchange"
const testQueue = "test_queue"
const testBindingKey = "binding-key"
const testAcceptanceBindingConfigMinimal = `
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

resource "qpid_queue" "test_queue" {
    name = "test_queue"
    depends_on = [qpid_virtual_host.acceptance_test_host]
    parents = ["acceptance_test", "acceptance_test_host"]
    type = "standard"
}

resource "qpid_exchange" "test_exchange" {
    name = "test_exchange"
    depends_on = [qpid_virtual_host.acceptance_test_host]
    parents = ["acceptance_test", "acceptance_test_host"]
    type = "direct"
}

resource "qpid_binding" "test_binding" {
    depends_on = [qpid_exchange.test_exchange, qpid_queue.test_queue]
    destination = "test_queue"
    exchange = "test_exchange"
    binding_key = "binding-key"
    parents = ["acceptance_test", "acceptance_test_host"]
    arguments = {
          "x-filter-jms-selector" = "foo='bar'"
    }
}
`
