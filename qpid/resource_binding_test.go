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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceBindingCheckDestroy(),
		Steps: []resource.TestStep{
			{
				// test binding created from configuration
				Config: testAcceptanceBindingConfigMinimal,
				Check: testAcceptanceBindingCheck(
					testAcceptanceBindingResource,
					&Binding{testBindingKey,
						testAcceptanceQueueName,
						testAcceptanceExchangeName,
						map[string]string{"x-filter-jms-selector": "foo='bar'"},
						testAcceptanceVirtualHostNodeName,
						testAcceptanceVirtualHostName},
				),
			},
			{
				// test binding updated from configuration
				Config: testAcceptanceBindingConfigArgumentsRemoved,
				Check: testAcceptanceBindingCheck(
					testAcceptanceBindingResource,
					&Binding{testBindingKey,
						testAcceptanceQueueName,
						testAcceptanceExchangeName,
						map[string]string{},
						testAcceptanceVirtualHostNodeName,
						testAcceptanceVirtualHostName},
				),
			},
			{
				// test broker side deleted binding is restored from configuration
				PreConfig: dropBinding(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName, testAcceptanceExchangeName, testAcceptanceQueueName, testBindingKey),
				Config:    testAcceptanceBindingConfigMinimal,
				Check: testAcceptanceBindingCheck(
					testAcceptanceBindingResource,
					&Binding{testBindingKey,
						testAcceptanceQueueName,
						testAcceptanceExchangeName,
						map[string]string{"x-filter-jms-selector": "foo='bar'"},
						testAcceptanceVirtualHostNodeName,
						testAcceptanceVirtualHostName},
				),
			},
			{
				// test binding is deleted after its deletion in configuration
				Config: testBindingParents,
				Check: testAcceptanceBindingDeleted(
					testAcceptanceBindingResource,
					&Binding{testBindingKey,
						testAcceptanceQueueName,
						testAcceptanceExchangeName,
						map[string]string{},
						testAcceptanceVirtualHostNodeName,
						testAcceptanceVirtualHostName},
				),
			},
		},
	})
}

func testAcceptanceBindingCheck(rn string, expected *Binding) resource.TestCheckFunc {
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

		if expected != nil && !reflect.DeepEqual(*bnd, *expected) {
			return fmt.Errorf("unexpected binding %s : expected %v, got %v", rn, bnd, expected)
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
		bindings, err := client.getExchangeBindings(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName, testAcceptanceExchangeName)
		if err != nil {
			return fmt.Errorf("error on getting bindings: %s", err)
		}

		for _, bnd := range *bindings {
			if bnd["bindingKey"] == testBindingKey && bnd["destination"] == testAcceptanceQueueName {
				return fmt.Errorf("binding %s/%s/%s/%s/%s still exist", testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName, testAcceptanceExchangeName, testAcceptanceQueueName, testBindingKey)
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

const testAcceptanceBindingName = "test_binding"
const testAcceptanceBindingResourceName = "qpid_binding"
const testAcceptanceBindingResource = "qpid_binding.test_binding"
const testBindingKey = "binding-key"
const testBindingParents = testAcceptanceExchangeConfigMinimal + testQueueConfiguration

const testAcceptanceBindingConfigMinimal = testBindingParents + `
resource "` + testAcceptanceBindingResourceName + `" "` + testAcceptanceBindingName + `" {
    depends_on = [` + testAcceptanceExchangeResource + `, ` + testAcceptanceQueueResource + `]
    destination = "` + testAcceptanceQueueName + `"
    exchange = "` + testAcceptanceExchangeName + `"
    binding_key = "` + testBindingKey + `"
    virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    arguments = {
          "x-filter-jms-selector" = "foo='bar'"
    }
}
`

const testAcceptanceBindingConfigArgumentsRemoved = testBindingParents + `
resource "` + testAcceptanceBindingResourceName + `" "` + testAcceptanceBindingName + `" {
    depends_on = [` + testAcceptanceExchangeResource + `, ` + testAcceptanceQueueResource + `]
    destination = "` + testAcceptanceQueueName + `"
    exchange = "` + testAcceptanceExchangeName + `"
    binding_key = "` + testBindingKey + `"
    virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
}
`
