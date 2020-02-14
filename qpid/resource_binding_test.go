package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"reflect"
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
				// test binding created from configuration
				Config: testAcceptanceBindingConfigMinimal,
				Check: testAcceptanceBindingCheck(
					"qpid_binding.test_binding", binding, &Binding{testBindingKey, testQueue, testExchangeName, map[string]string{"x-filter-jms-selector": "foo='bar'"}, testNodeName, testHostName},
				),
			},
			{
				// test binding updated from configuration
				Config: testAcceptanceBindingConfigArgumentsRemoved,
				Check: testAcceptanceBindingCheck(
					"qpid_binding.test_binding", binding, &Binding{testBindingKey, testQueue, testExchangeName, map[string]string{}, testNodeName, testHostName},
				),
			},
			{
				// test broker side deleted binding is restored from configuration
				PreConfig: dropBinding(testNodeName, testHostName, testExchangeName, testQueue, testBindingKey),
				Config:    testAcceptanceBindingConfigMinimal,
				Check: testAcceptanceBindingCheck(
					"qpid_binding.test_binding", binding, &Binding{testBindingKey, testQueue, testExchangeName, map[string]string{"x-filter-jms-selector": "foo='bar'"}, testNodeName, testHostName},
				),
			},
			{
				// test binding is deleted after its deletion in configuration
				Config: testBindingParents,
				Check: testAcceptanceBindingDeleted(
					"qpid_binding.test_binding", &Binding{testBindingKey, testQueue, testExchangeName, map[string]string{}, testNodeName, testHostName},
				),
			},
		},
	})
}

func testAcceptanceBindingCheck(rn string, binding *Binding, expected *Binding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("binding id not set")
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

		bnd, err := findBinding(node, host, exch, dest, key)
		if err != nil {
			return fmt.Errorf("error on getting binding: %s", err)
		}

		if bnd == nil {
			return fmt.Errorf("unable to find binding %s", rn)
		}

		binding = bnd
		if expected != nil && !reflect.DeepEqual(*binding, *expected) {
			return fmt.Errorf("unexpected binding %s : expected %v, got %v", rn, binding, expected)
		}
		return nil
	}
}

func testAcceptanceBindingDeleted(rn string, binding *Binding) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[rn]
		if ok {
			return fmt.Errorf("deleted resource found: %s", rn)
		}

		bnd, err := findBinding((*binding).VirtualHostNode, (*binding).VirtualHost, (*binding).Exchange, (*binding).Destination, (*binding).BindingKey)
		if err != nil {
			return fmt.Errorf("error on getting binding: %s", err)
		}
		if bnd != nil {
			return fmt.Errorf("found deleted binding : %v", *binding)
		}
		return nil
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

func findBinding(node string, host string, exchange string, destination string, bindingKey string) (*Binding, error) {
	client := testAcceptanceProvider.Meta().(*Client)
	return client.GetBinding(&Binding{bindingKey, destination, exchange, nil, node, host})
}

const testNodeName = "acceptance_test"
const testHostName = "acceptance_test_host"
const testExchangeName = "test_exchange"
const testQueue = "test_queue"
const testBindingKey = "binding-key"
const testBindingParents = `
resource "qpid_virtual_host_node" "acceptance_test" {
    name = "acceptance_test"
    type = "JSON"
    virtual_host_initial_configuration = "{}"
}

resource "qpid_virtual_host" "acceptance_test_host" {
    depends_on = [qpid_virtual_host_node.acceptance_test]
    name = "acceptance_test_host"
    virtual_host_node = "acceptance_test"
    type = "BDB"
}

resource "qpid_queue" "test_queue" {
    name = "test_queue"
    depends_on = [qpid_virtual_host.acceptance_test_host]
    virtual_host_node = "acceptance_test"
    virtual_host = "acceptance_test_host"
    type = "standard"
}

resource "qpid_exchange" "test_exchange" {
    name = "test_exchange"
    depends_on = [qpid_virtual_host.acceptance_test_host]
    virtual_host_node = "acceptance_test"
    virtual_host = "acceptance_test_host"
    type = "direct"
}
`

const testAcceptanceBindingConfigMinimal = testBindingParents + `
resource "qpid_binding" "test_binding" {
    depends_on = [qpid_exchange.test_exchange, qpid_queue.test_queue]
    destination = "test_queue"
    exchange = "test_exchange"
    binding_key = "binding-key"
    virtual_host_node = "acceptance_test"
    virtual_host = "acceptance_test_host"
    arguments = {
          "x-filter-jms-selector" = "foo='bar'"
    }
}
`

const testAcceptanceBindingConfigArgumentsRemoved = testBindingParents + `
resource "qpid_binding" "test_binding" {
    depends_on = [qpid_exchange.test_exchange, qpid_queue.test_queue]
    destination = "test_queue"
    exchange = "test_exchange"
    binding_key = "binding-key"
    virtual_host_node = "acceptance_test"
    virtual_host = "acceptance_test_host"
}
`
