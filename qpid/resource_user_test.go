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
			return fmt.Errorf("parent not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)

		hosts, err := client.GetUsers(authenticationProvider)
		if err != nil {
			return fmt.Errorf("error on getting hosts: %s", err)
		}

		for _, host := range *hosts {
			if host["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&host, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("unable to find virtualhost %s", rn)
	}
}

func testAcceptanceUserCheckDestroy(testAuthenticationProviderName string, virtualHostName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)

		hosts, err := client.GetUsers(testAuthenticationProviderName)
		if err != nil {
			return fmt.Errorf("error on getting users for node '%s' : %s", virtualHostName, err)
		}

		for _, host := range *hosts {
			if host["name"] == virtualHostName {
				return fmt.Errorf("user %s/%s still exist", testAuthenticationProviderName, virtualHostName)
			}
		}

		return nil
	}
}

func dropUser(nodeName string, hostName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteUser(nodeName, hostName)
		if err != nil {
			fmt.Printf("unable to delete user : %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete user: %v", resp))
		}
	}
}

const testAcceptanceUserName = "test_user"
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
