package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceUser(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceUserCheckDestroy(testAcceptanceAuthenticationProviderName, testAcceptanceUserName),
		Steps: []resource.TestStep{
			{
				Config: testAcceptanceUserConfigMinimal,
				Check: testAcceptanceUserCheck(
					testAcceptanceUserResource,
					&map[string]interface{}{"name": testAcceptanceUserName, "type": "managed"},
				),
			},
			{
				PreConfig: dropUser(testAcceptanceAuthenticationProviderName, testAcceptanceUserName),
				Config:    testAcceptanceUserConfigMinimal,
				Check: testAcceptanceUserCheck(
					testAcceptanceUserResource,
					&map[string]interface{}{"name": testAcceptanceUserName, "type": "managed"},
				),
			},

			{
				Config: testAcceptanceUserParent + `

resource "` + testAcceptanceUserResourceName + `" "` + testAcceptanceUserName + `" {
    depends_on = [` + testAcceptanceAuthenticationProviderResource + `]
    name = "` + testAcceptanceUserName + `"
    authentication_provider = "` + testAcceptanceAuthenticationProviderName + `"
    type = "managed"
    password = "foo"
}`,
				Check: testAcceptanceUserCheck(
					testAcceptanceUserResource,
					&map[string]interface{}{"name": testAcceptanceUserName, "type": "managed"},
				),
			},
		},
	})
}

func testAcceptanceUserCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("user id not set")
		}

		authenticationProvider, ok := rs.Primary.Attributes["authentication_provider"]
		if !ok {
			return fmt.Errorf("authentication_provider not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		users, err := client.GetUsers(authenticationProvider)
		if err != nil {
			return fmt.Errorf("error on getting users: %s", err)
		}

		for _, user := range *users {
			if user["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&user, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find user %s", rn)
	}
}

func testAcceptanceUserCheckDestroy(testAuthenticationProviderName string, virtualHostName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		users, err := client.GetUsers(testAuthenticationProviderName)
		if err != nil {
			return fmt.Errorf("error on getting users for node '%s' : %s", virtualHostName, err)
		}

		for _, user := range *users {
			if user["name"] == virtualHostName {
				return fmt.Errorf("user %s/%s still exist", testAuthenticationProviderName, virtualHostName)
			}
		}

		return nil
	}
}

func dropUser(providerName string, userName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteUser(providerName, userName)
		if err != nil {
			fmt.Printf("unable to delete user : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete user: %v", resp))
		}
	}
}

const testAcceptanceUserName = "acceptance_test_user"
const testAcceptanceUserResourceName = "qpid_user"
const testAcceptanceUserResource = testAcceptanceUserResourceName + "." + testAcceptanceUserName

const testAcceptanceUserParent = testAcceptanceAuthenticationProviderPlainConfigMinimal

const testAcceptanceUserConfigMinimal = testAcceptanceUserParent + `
resource "` + testAcceptanceUserResourceName + `" "` + testAcceptanceUserName + `" {
    depends_on = [` + testAcceptanceAuthenticationProviderResource + `]
    name = "` + testAcceptanceUserName + `"
    authentication_provider = "` + testAcceptanceAuthenticationProviderName + `"
    type = "managed"
    password = "password"
}`
