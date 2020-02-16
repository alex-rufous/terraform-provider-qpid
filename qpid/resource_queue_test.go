package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceQueue(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceQueueCheckDestroy(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName, testAcceptanceQueueName),
		Steps: []resource.TestStep{
			{
				// test new queue creation from configuration
				Config: testAcceptanceQueueConfigMinimal,
				Check: testAcceptanceQueueCheck(
					testAcceptanceQueueResource, &map[string]interface{}{"name": testAcceptanceQueueName, "type": "standard"},
				),
			},
			{
				// test queue restoration from configuration after its deletion on broker side
				PreConfig: dropQueue(testAcceptanceVirtualHostNodeName, testAcceptanceVirtualHostName, testAcceptanceQueueName),
				Config:    testAcceptanceQueueConfigMinimal,
				Check: testAcceptanceQueueCheck(
					testAcceptanceQueueResource, &map[string]interface{}{"name": testAcceptanceQueueName, "type": "standard"},
				),
			},
			{
				// update queue attributes
				Config: getQueueConfigurationWithAttributes(&map[string]string{"minimum_message_ttl": "1000", "maximum_message_ttl": "99999"}),
				Check: testAcceptanceQueueCheck(
					testAcceptanceQueueResource, &map[string]interface{}{"minimumMessageTtl": 1000.0, "maximumMessageTtl": 99999.0},
				),
			},
			{
				// update with alternate binding
				Config: testAcceptanceVirtualHostConfigMinimal + testAcceptanceQueue2 + `
resource "` + testAcceptanceQueueResourceName + `" "` + testAcceptanceQueueName + `" {
    name = "` + testAcceptanceQueueName + `"
    depends_on = [` + testAcceptanceVirtualHostResource + `, ` + testAcceptanceQueueResource2 + `]
	virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    type = "standard"
    alternate_binding {
		destination = "` + testAcceptanceQueueName2 + `"
        attributes = {
						"x-filter-jms-selector"= "id>0"
        }
    }
}
`,
				Check: testAcceptanceQueueCheck(
					testAcceptanceQueueResource, &map[string]interface{}{
						"alternateBinding": map[string]interface{}{
							"destination": testAcceptanceQueueName2,
							"attributes": map[string]interface{}{
								"x-filter-jms-selector": "id>0",
							}}},
					"minimumMessageTtl",
					"maximumMessageTtl",
				),
			},

			{
				// update with removed attributes
				Config: getQueueConfigurationWithAttributes(&map[string]string{"message_durability": "\"NEVER\""}) + testAcceptanceQueue2,
				Check: testAcceptanceQueueCheck(
					testAcceptanceQueueResource,
					&map[string]interface{}{"messageDurability": "NEVER"},
					"alternateBinding",
				),
			},
		},
	})
}

func testAcceptanceQueueCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("queue id not set")
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

		queues, err := client.getVirtualHostQueues(nodeName, hostName)
		if err != nil {
			return fmt.Errorf("error on getting queues: %s", err)
		}

		for _, queue := range *queues {
			if queue["id"] != nil && queue["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&queue, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find queue %s", rn)
	}
}

func testAcceptanceQueueCheckDestroy(virtualHostNodeName string, virtualHostName string, queueName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		queues, err := client.getVirtualHostQueues(virtualHostNodeName, virtualHostName)
		if err != nil {
			return fmt.Errorf("error on getting queues: %s", err)
		}

		for _, queue := range *queues {
			if queue["name"] == queueName {
				return fmt.Errorf("queue %s/%s/%s still exist", virtualHostNodeName, virtualHostName, queueName)
			}
		}

		return nil
	}
}

func dropQueue(nodeName string, hostName string, queueName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteQueue(nodeName, hostName, queueName)
		if err != nil {
			fmt.Printf("unable to delete queue : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete queue: %v", resp))
		}
	}
}

const testAcceptanceQueueName = "test_queue"
const testAcceptanceQueueName2 = "test_queue2"
const testAcceptanceQueueResourceName = "qpid_queue"
const testAcceptanceQueueResource = testAcceptanceQueueResourceName + "." + testAcceptanceQueueName
const testAcceptanceQueueResource2 = testAcceptanceQueueResourceName + "." + testAcceptanceQueueName2
const testQueueConfiguration = `
resource "` + testAcceptanceQueueResourceName + `" "` + testAcceptanceQueueName + `" {
    name = "` + testAcceptanceQueueName + `"
    depends_on = [` + testAcceptanceVirtualHostResource + `]
	virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    type = "standard"
}
`
const testAcceptanceQueueConfigMinimal = testAcceptanceVirtualHostConfigMinimal + testQueueConfiguration
const testAcceptanceQueue2 = `
resource "` + testAcceptanceQueueResourceName + `" "` + testAcceptanceQueueName2 + `" {
    name = "` + testAcceptanceQueueName2 + `"
    depends_on = [` + testAcceptanceVirtualHostResource + `]
	virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    type = "standard"
}
`

func getQueueConfigurationWithAttributes(entries *map[string]string) string {
	config := testAcceptanceVirtualHostConfigMinimal + `
resource "` + testAcceptanceQueueResourceName + `" "` + testAcceptanceQueueName + `" {
    name = "` + testAcceptanceQueueName + `"
    depends_on = [` + testAcceptanceVirtualHostResource + `]
	virtual_host_node = "` + testAcceptanceVirtualHostNodeName + `"
    virtual_host = "` + testAcceptanceVirtualHostName + `"
    type = "standard"
`
	for k, v := range *entries {
		config += fmt.Sprintf("    %s=%s\n", k, v)
	}
	config += `}
`
	return config
}
