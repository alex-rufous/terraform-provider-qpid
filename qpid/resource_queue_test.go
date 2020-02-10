package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceQueue(t *testing.T) {
	var virtualHostNodeName string
	var virtualHostName string
	var queueName string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceQueueCheckDestroy(virtualHostNodeName, virtualHostName, queueName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceQueueConfigMinimal,
				Check: testAcceptanceQueueCheck(
					"qpid_queue.test_queue", &virtualHostNodeName, &virtualHostName, &queueName,
				),
			},
			{
				PreConfig: dropQueue("acceptance_test", "acceptance_test_host", "test_queue"),
				Config:    testAcceptanceQueueConfigMinimal,
				Check: testAcceptanceQueueCheck(
					"qpid_queue.test_queue", &virtualHostNodeName, &virtualHostName, &queueName,
				),
			},
		},
	})
}

func testAcceptanceQueueCheck(rn string, virtualHostNodeName *string, virtualHostName *string, queueName *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("queue id not set")
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

		queues, err := client.getVirtualHostQueues(nodeName, hostName)
		if err != nil {
			return fmt.Errorf("error on getting queues: %s", err)
		}

		for _, queue := range *queues {
			if queue["id"] == rs.Primary.ID {
				*queueName = queue["name"].(string)
				*virtualHostNodeName = nodeName
				*virtualHostName = hostName
				return nil
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

const testAcceptanceQueueConfigMinimal = `
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
`
