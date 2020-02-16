package qpid

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAcceptanceExchange(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceExchangeCheckDestroy(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName, testAcceptanceExchangeName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceExchangeConfigMinimal,
				Check: testAcceptanceExchangeCheck(
					testAcceptanceExchangeResource, &map[string]interface{}{"name": testAcceptanceExchangeName, "type": "direct"},
				),
			},
			{
				PreConfig: dropExchange(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName, testAcceptanceExchangeName),
				Config:    testAcceptanceExchangeConfigMinimal,
				Check: testAcceptanceExchangeCheck(
					testAcceptanceExchangeResource, &map[string]interface{}{"name": testAcceptanceExchangeName, "type": "direct"},
				),
			},
			{
				// update with alternate binding
				Config: testAcceptanceVirtualHostConfigMinimal + testAcceptanceQueue2 + `
resource "` + testAcceptanceExchangeResourceName + `" "` + testAcceptanceExchangeName + `" {
    name = "` + testAcceptanceExchangeName + `"
    depends_on = [` + testAcceptanceVirtualHostResource + `, ` + testAcceptanceQueueResource2 + `]
	virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    type = "direct"
    alternate_binding {
		destination = "` + testAcceptanceQueueName2 + `"
        attributes = {
						"x-filter-jms-selector"= "id>0"
        }
    }
}
`,
				Check: testAcceptanceExchangeCheck(
					testAcceptanceExchangeResource, &map[string]interface{}{
						"alternateBinding": map[string]interface{}{
							"destination": testAcceptanceQueueName2,
							"attributes": map[string]interface{}{
								"x-filter-jms-selector": "id>0",
							}}},
				),
			},

			{
				// update with removed attributes
				Config: getExchangeConfigurationWithAttributes(&map[string]string{"unroutable_message_behaviour": "\"REJECT\""}) + testAcceptanceQueue2,
				Check: testAcceptanceExchangeCheck(
					testAcceptanceExchangeResource,
					&map[string]interface{}{"unroutableMessageBehaviour": "REJECT"},
					"alternateBinding",
				),
			},
		},
	})
}

func testAcceptanceExchangeCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("exchange id not set")
		}

		nodeName, ok := rs.Primary.Attributes["virtual_host_node"]
		if !ok {
			return fmt.Errorf("node name is not set")
		}

		hostName, ok := rs.Primary.Attributes["virtual_host"]
		if !ok {
			return fmt.Errorf("host name is not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		exchanges, err := client.getVirtualHostExchanges(nodeName, hostName)
		if err != nil {
			return fmt.Errorf("error on getting exchanges: %s", err)
		}

		for _, exchange := range *exchanges {
			if exchange["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&exchange, expectedAttributes, removed)
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

const testAcceptanceExchangeName = "test_exchange"
const testAcceptanceExchangeResourceName = "qpid_exchange"
const testAcceptanceExchangeResource = testAcceptanceExchangeResourceName + "." + testAcceptanceExchangeName
const testAcceptanceExchangeConfigMinimal = testAcceptanceVirtualHostConfigMinimal + `

resource "` + testAcceptanceExchangeResourceName + `" "` + testAcceptanceExchangeName + `" {
    name = "` + testAcceptanceExchangeName + `"
    depends_on = [` + testAcceptanceVirtualHostResource + `]
    virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    type = "direct"
}
`

func getExchangeConfigurationWithAttributes(entries *map[string]string) string {
	config := testAcceptanceVirtualHostConfigMinimal + `
resource "` + testAcceptanceExchangeResourceName + `" "` + testAcceptanceExchangeName + `" {
    name = "` + testAcceptanceExchangeName + `"
    depends_on = [` + testAcceptanceVirtualHostResource + `]
	virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    type = "direct"
`
	for k, v := range *entries {
		config += fmt.Sprintf("    %s=%s\n", k, v)
	}
	config += "}\n"

	return config
}
