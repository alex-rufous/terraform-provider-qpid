package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceGroup(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceGroupCheckDestroy(testAcceptanceGroupProviderName, testAcceptanceGroupName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceGroupConfigMinimal,
				Check: testAcceptanceGroupCheck(
					testAcceptanceGroupResource,
					&map[string]interface{}{"name": testAcceptanceGroupName, "type": "ManagedGroup"},
				),
			},
			{
				PreConfig: dropGroup(testAcceptanceGroupProviderName, testAcceptanceGroupName),
				Config:    testAcceptanceGroupConfigMinimal,
				Check: testAcceptanceGroupCheck(
					testAcceptanceGroupResource,
					&map[string]interface{}{"name": testAcceptanceGroupName, "type": "ManagedGroup"},
				),
			},

			{
				Config: testAcceptanceGroupProviderManagedGroupProviderConfigMinimal + `

resource "` + testAcceptanceGroupResourceName + `" "` + testAcceptanceGroupName + `" {
    depends_on = [` + testAcceptanceGroupProviderResource + `]
    name = "` + testAcceptanceGroupName + `"
    group_provider = "` + testAcceptanceGroupProviderName + `"
    type = "ManagedGroup"
    description = "Test group"
}`,
				Check: testAcceptanceGroupCheck(
					testAcceptanceGroupResource,
					&map[string]interface{}{"name": testAcceptanceGroupName, "type": "ManagedGroup", "description": "Test group"},
				),
			},
		},
	})
}

func testAcceptanceGroupCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("group id not set")
		}

		groupProvider, ok := rs.Primary.Attributes["group_provider"]
		if !ok {
			return fmt.Errorf("group_provider not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		groups, err := client.GetGroups(groupProvider)

		if err != nil {
			return fmt.Errorf("error on getting groups: %s", err)
		}

		for _, group := range *groups {
			if group["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&group, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find group %s", rn)
	}
}

func testAcceptanceGroupCheckDestroy(testGroupProviderName string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		groups, err := client.GetGroups(testGroupProviderName)
		if err != nil {
			return fmt.Errorf("error on getting groups for group provider '%s' : %s", testGroupProviderName, err)
		}

		for _, group := range *groups {
			if group["name"] == name {
				return fmt.Errorf("group %s/%s still exist", testGroupProviderName, name)
			}
		}

		return nil
	}
}

func dropGroup(providerName string, groupName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteGroup(providerName, groupName)
		if err != nil {
			fmt.Printf("unable to delete group : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete group: %v", resp))
		}
	}
}

const testAcceptanceGroupName = "acceptance_test_group"
const testAcceptanceGroupResourceName = "qpid_group"
const testAcceptanceGroupResource = testAcceptanceGroupResourceName + "." + testAcceptanceGroupName
const testAcceptanceGroupConfigMinimal = testAcceptanceGroupProviderManagedGroupProviderConfigMinimal + `
resource "` + testAcceptanceGroupResourceName + `" "` + testAcceptanceGroupName + `" {
    depends_on = [` + testAcceptanceGroupProviderResource + `]
    name = "` + testAcceptanceGroupName + `"
    group_provider = "` + testAcceptanceGroupProviderName + `"
    type = "ManagedGroup"
}`
