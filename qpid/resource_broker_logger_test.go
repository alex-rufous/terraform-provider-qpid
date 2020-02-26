package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceFileBrokerLogger(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceBrokerLoggerCheckDestroy(testAcceptanceBrokerLoggerName),
		Steps: []resource.TestStep{
			{
				// test new broker logger creation from configuration
				Config: testAcceptanceFileBrokerLoggerConfigMinimal,
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName, "type": "File"},
				),
			},
			{
				// test broker logger restoration from configuration after its deletion on broker side
				PreConfig: dropBrokerLogger(testAcceptanceBrokerLoggerName),
				Config:    testAcceptanceFileBrokerLoggerConfigMinimal,
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName, "type": "File"},
				),
			},
			{
				// test broker logger update
				Config: getBrokerLoggerConfigurationWithAttributes("File", "compress_old_files = true"),
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName,
						"type":             "File",
						"compressOldFiles": true},
				),
			},
			{
				// test broker logger attribute removal
				Config: testAcceptanceFileBrokerLoggerConfigMinimal,
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName, "type": "File"},
					"compressOldFiles",
				),
			},
		},
	})
}

func TestAcceptanceConsoleBrokerLogger(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceBrokerLoggerCheckDestroy(testAcceptanceBrokerLoggerName),
		Steps: []resource.TestStep{
			{
				// test new broker logger creation from configuration
				Config: getBrokerLoggerConfigurationWithAttributes("Console"),
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName, "type": "Console"},
				),
			},

			{
				// test broker logger restoration from configuration after its deletion on broker side
				PreConfig: dropBrokerLogger(testAcceptanceBrokerLoggerName),
				Config:    getBrokerLoggerConfigurationWithAttributes("Console"),
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName, "type": "Console"},
				),
			},
			{
				// test broker logger update
				Config: getBrokerLoggerConfigurationWithAttributes("Console", "console_stream_target = \"STDERR\""),
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName,
						"type":                "Console",
						"consoleStreamTarget": "STDERR"},
				),
			},

			{
				// test broker logger attribute removal
				Config: getBrokerLoggerConfigurationWithAttributes("Console"),
				Check: testAcceptanceBrokerLoggerCheck(
					testAcceptanceBrokerLoggerResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerName, "type": "Console"},
					"consoleStreamTarget",
				),
			},
		},
	})
}
func dropBrokerLogger(nodeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteBrokerLogger(nodeName)
		if err != nil {
			fmt.Printf("unable to delete broker logger: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete broker logger: %v", resp))
		}
	}
}

func testAcceptanceBrokerLoggerCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("broker logger id not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)
		providers, err := client.GetBrokerLoggers()
		if err != nil {
			return fmt.Errorf("error getting broker logger: %s", err)
		}

		for _, provider := range *providers {
			if provider["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&provider, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("broker logger '%s' is not found", rn)
	}
}

func testAcceptanceBrokerLoggerCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		loggers, err := client.GetBrokerLoggers()
		if err != nil {
			return fmt.Errorf("error getting loggers: %s", err)
		}

		for _, logger := range *loggers {
			if logger["name"] == name {
				return fmt.Errorf("broker logger '%v' still exists", logger)
			}
		}

		return nil
	}
}

const testAcceptanceBrokerLoggerResourceName = "qpid_broker_logger"
const testAcceptanceBrokerLoggerName = "acceptance_test_broker_logger"
const testAcceptanceBrokerLoggerResource = testAcceptanceBrokerLoggerResourceName + "." + testAcceptanceBrokerLoggerName

const testAcceptanceFileBrokerLoggerConfigMinimal = testAcceptanceAuthenticationProviderPlainConfigMinimal + `
resource "` + testAcceptanceBrokerLoggerResourceName + `" "` + testAcceptanceBrokerLoggerName + `" {
    name = "` + testAcceptanceBrokerLoggerName + `"
    type = "File"
    file_name = "$${qpid.work_dir}$${file.separator}log$${file.separator}acceptance_test_broker_logger.log"
}
`

func getBrokerLoggerConfigurationWithAttributes(typeName string, entries ...string) string {
	config := testAcceptanceAuthenticationProviderPlainConfigMinimal + `
resource "` + testAcceptanceBrokerLoggerResourceName + `" "` + testAcceptanceBrokerLoggerName + `" {
    name = "` + testAcceptanceBrokerLoggerName + `"
    type = "` + typeName + `"
`

	for _, v := range entries {
		config += fmt.Sprintf("    %v\n", v)
	}
	config += `}
`
	return config
}
