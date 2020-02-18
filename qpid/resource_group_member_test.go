package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceGroupMember(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceGroupMemberCheckDestroy(testAcceptanceGroupProviderName, testAcceptanceGroupName, testAcceptanceGroupMemberName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceGroupMemberConfigMinimal,
				Check: testAcceptanceGroupMemberCheck(
					testAcceptanceGroupMemberResource,
					&map[string]interface{}{"name": testAcceptanceGroupMemberName, "type": "ManagedGroupMember"},
				),
			},
			{
				PreConfig: dropGroupMember(testAcceptanceGroupProviderName, testAcceptanceGroupName, testAcceptanceGroupMemberName),
				Config:    testAcceptanceGroupMemberConfigMinimal,
				Check: testAcceptanceGroupMemberCheck(
					testAcceptanceGroupMemberResource,
					&map[string]interface{}{"name": testAcceptanceGroupMemberName, "type": "ManagedGroupMember"},
				),
			},

			{
				Config: testAcceptanceGroupConfigMinimal + `

resource "` + testAcceptanceGroupMemberResourceName + `" "` + testAcceptanceGroupMemberName + `" {
    depends_on = [` + testAcceptanceGroupProviderResource + `,` + testAcceptanceGroupResource + `]
    name = "` + testAcceptanceGroupMemberName + `"
    group_provider = "` + testAcceptanceGroupProviderName + `"
	group = "` + testAcceptanceGroupName + `"
    type = "ManagedGroupMember"
    description = "Test group member"
}`,
				Check: testAcceptanceGroupMemberCheck(
					testAcceptanceGroupMemberResource,
					&map[string]interface{}{"name": testAcceptanceGroupMemberName, "type": "ManagedGroupMember", "description": "Test group member"},
				),
			},
		},
	})
}

func testAcceptanceGroupMemberCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("group member id not set")
		}

		groupProvider, ok := rs.Primary.Attributes["group_provider"]
		if !ok {
			return fmt.Errorf("group_provider not set")
		}

		groupName, ok := rs.Primary.Attributes["group"]
		if !ok {
			return fmt.Errorf("group not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		members, err := client.GetGroupMembers(groupProvider, groupName)

		if err != nil {
			return fmt.Errorf("error on getting group members: %s", err)
		}

		for _, member := range *members {
			if member["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&member, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find group member %s", rn)
	}
}

func testAcceptanceGroupMemberCheckDestroy(providerName string, groupName string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		members, err := client.GetGroupMembers(providerName, groupName)
		if err != nil {
			return fmt.Errorf("error on getting group members for group '%s'/%s : %s", providerName, groupName, err)
		}

		for _, member := range *members {
			if member["name"] == name {
				return fmt.Errorf("group member %s/%s/%s still exist", providerName, groupName, name)
			}
		}

		return nil
	}
}

func dropGroupMember(providerName string, groupName string, name string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteGroupMember(providerName, groupName, name)
		if err != nil {
			fmt.Printf("unable to delete group member: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete group member: %v", resp))
		}
	}
}

const testAcceptanceGroupMemberName = "acceptance_test_group_member"
const testAcceptanceGroupMemberResourceName = "qpid_group_member"
const testAcceptanceGroupMemberResource = testAcceptanceGroupMemberResourceName + "." + testAcceptanceGroupMemberName
const testAcceptanceGroupMemberConfigMinimal = testAcceptanceGroupConfigMinimal + `
resource "` + testAcceptanceGroupMemberResourceName + `" "` + testAcceptanceGroupMemberName + `" {
    depends_on = [` + testAcceptanceGroupProviderResource + `,` + testAcceptanceGroupResource + `]
    name = "` + testAcceptanceGroupMemberName + `"
    group_provider = "` + testAcceptanceGroupProviderName + `"
	group = "` + testAcceptanceGroupName + `"
    type = "ManagedGroupMember"
}`
