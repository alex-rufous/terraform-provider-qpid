package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceBrokerLoggerRule(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceBrokerLoggerRuleCheckDestroy(testAcceptanceBrokerLoggerName, testAcceptanceBrokerLoggerRuleName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceBrokerLoggerRuleConfigMinimal,
				Check: testAcceptanceBrokerLoggerRuleCheck(
					testAcceptanceBrokerLoggerRuleResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerRuleName, "type": "NameAndLevel", "level": "INFO"},
				),
			},

			{
				PreConfig: dropBrokerLoggerRule(testAcceptanceBrokerLoggerName, testAcceptanceBrokerLoggerRuleName),
				Config:    testAcceptanceBrokerLoggerRuleConfigMinimal,
				Check: testAcceptanceBrokerLoggerRuleCheck(
					testAcceptanceBrokerLoggerRuleResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerRuleName, "type": "NameAndLevel", "level": "INFO"},
				),
			},
			{
				Config: testAcceptanceFileBrokerLoggerConfigMinimal + `

					resource "` + testAcceptanceBrokerLoggerRuleResourceName + `" "` + testAcceptanceBrokerLoggerRuleName + `" {
					    depends_on = [` + testAcceptanceBrokerLoggerResource + `]
					    name = "` + testAcceptanceBrokerLoggerRuleName + `"
					    broker_logger = "` + testAcceptanceBrokerLoggerName + `"
					    type = "NameAndLevel"
					    level = "ERROR"
					    logger_name = "org.apache.*"
					}`,
				Check: testAcceptanceBrokerLoggerRuleCheck(
					testAcceptanceBrokerLoggerRuleResource,
					&map[string]interface{}{"name": testAcceptanceBrokerLoggerRuleName, "type": "NameAndLevel", "level": "ERROR", "loggerName": "org.apache.*"},
				),
			},
		},
	})
}

func testAcceptanceBrokerLoggerRuleCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("rule id not set")
		}

		brokerLogger, ok := rs.Primary.Attributes["broker_logger"]
		if !ok {
			return fmt.Errorf("broker logger not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		rules, err := client.GetBrokerLoggerRules(brokerLogger)

		if err != nil {
			return fmt.Errorf("error on getting rules: %s", err)
		}

		for _, rule := range *rules {
			if rule["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&rule, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find rule %s", rn)
	}
}

func testAcceptanceBrokerLoggerRuleCheckDestroy(loggerName string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		rules, err := client.GetBrokerLoggerRules(loggerName)
		if err != nil {
			return fmt.Errorf("error on getting rules for logger '%s' : %s", loggerName, err)
		}

		for _, rule := range *rules {
			if rule["name"] == name {
				return fmt.Errorf("rule %s/%s still exist", loggerName, name)
			}
		}

		return nil
	}
}

func dropBrokerLoggerRule(providerName string, ruleName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteBrokerLoggerRule(providerName, ruleName)
		if err != nil {
			fmt.Printf("unable to delete rule : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete rule: %v", resp))
		}
	}
}

const testAcceptanceBrokerLoggerRuleName = "acceptance_test_broker_logger_rule"
const testAcceptanceBrokerLoggerRuleResourceName = "qpid_broker_logger_rule"
const testAcceptanceBrokerLoggerRuleResource = testAcceptanceBrokerLoggerRuleResourceName + "." + testAcceptanceBrokerLoggerRuleName
const testAcceptanceBrokerLoggerRuleConfigMinimal = testAcceptanceFileBrokerLoggerConfigMinimal + `
resource "` + testAcceptanceBrokerLoggerRuleResourceName + `" "` + testAcceptanceBrokerLoggerRuleName + `" {
    depends_on = [` + testAcceptanceBrokerLoggerResource + `]
    name = "` + testAcceptanceBrokerLoggerRuleName + `"
    broker_logger = "` + testAcceptanceBrokerLoggerName + `"
    type = "NameAndLevel"
    level = "INFO"
    logger_name = "org.apache.*"
}`
