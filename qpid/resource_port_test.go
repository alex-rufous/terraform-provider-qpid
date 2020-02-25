package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceAmqpPort(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptancePortCheckDestroy(testAcceptancePortName),
		Steps: []resource.TestStep{
			{
				// test new port creation from configuration
				Config: testAcceptanceAmqpPortConfigMinimal,
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName, "type": "AMQP"},
				),
			},
			{
				// test port restoration from configuration after its deletion on broker side
				PreConfig: dropPort(testAcceptancePortName),
				Config:    testAcceptanceAmqpPortConfigMinimal,
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName, "type": "AMQP"},
				),
			},
			{
				// test port update
				Config: getPortConfigurationWithAttributes("AMQP", "allow_confidential_operations_on_insecure_channels = true"),
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName,
						"type": "AMQP",
						"allowConfidentialOperationsOnInsecureChannels": true},
				),
			},
			{
				// test port attribute removal
				Config: testAcceptanceAmqpPortConfigMinimal,
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName, "type": "AMQP"},
					"allowConfidentialOperationsOnInsecureChannels",
				),
			},
		},
	})
}

func TestAcceptanceHttpPort(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptancePortCheckDestroy(testAcceptancePortName),
		Steps: []resource.TestStep{
			{
				// test new port creation from configuration
				Config: getPortConfigurationWithAttributes("HTTP"),
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName, "type": "HTTP"},
				),
			},

			{
				// test port restoration from configuration after its deletion on broker side
				PreConfig: dropPort(testAcceptancePortName),
				Config:    getPortConfigurationWithAttributes("HTTP"),
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName, "type": "HTTP"},
				),
			},
			{
				// test port update
				Config: getPortConfigurationWithAttributes("HTTP", "thread_pool_maximum = 10"),
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName,
						"type":              "HTTP",
						"threadPoolMaximum": 10.0},
				),
			},

			{
				// test port attribute removal
				Config: getPortConfigurationWithAttributes("HTTP"),
				Check: testAcceptancePortCheck(
					testAcceptancePortResource,
					&map[string]interface{}{"name": testAcceptancePortName, "type": "HTTP"},
					"threadPoolMaximum",
				),
			},
		},
	})
}
func dropPort(nodeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeletePort(nodeName)
		if err != nil {
			fmt.Printf("unable to delete port: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete port: %v", resp))
		}
	}
}

func testAcceptancePortCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("port id not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)
		providers, err := client.GetPorts()
		if err != nil {
			return fmt.Errorf("error getting port: %s", err)
		}

		for _, provider := range *providers {
			if provider["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&provider, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("port '%s' is not found", rn)
	}
}

func testAcceptancePortCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		providers, err := client.GetPorts()
		if err != nil {
			return fmt.Errorf("error getting providers: %s", err)
		}

		for _, node := range *providers {
			if node["name"] == name {
				return fmt.Errorf("port '%v' still exists", node)
			}
		}

		return nil
	}
}

const testAcceptancePortResourceName = "qpid_port"
const testAcceptancePortName = "acceptance_test_port"
const testAcceptancePortResource = testAcceptancePortResourceName + "." + testAcceptancePortName

const testAcceptanceAmqpPortConfigMinimal = testAcceptanceAuthenticationProviderPlainConfigMinimal + `
resource "` + testAcceptancePortResourceName + `" "` + testAcceptancePortName + `" {
    depends_on=[` + testAcceptanceAuthenticationProviderResource + `]
    name = "` + testAcceptancePortName + `"
    type = "AMQP"
    port = 0
    authentication_provider="` + testAcceptanceAuthenticationProviderName + `"
}
`

func getPortConfigurationWithAttributes(typeName string, entries ...string) string {
	config := testAcceptanceAuthenticationProviderPlainConfigMinimal + `
resource "` + testAcceptancePortResourceName + `" "` + testAcceptancePortName + `" {
    depends_on=[` + testAcceptanceAuthenticationProviderResource + `]
    name = "` + testAcceptancePortName + `"
    type = "` + typeName + `"
    port = 0
    authentication_provider="` + testAcceptanceAuthenticationProviderName + `"
`

	for _, v := range entries {
		config += fmt.Sprintf("    %v\n", v)
	}
	config += `}
`
	return config
}
