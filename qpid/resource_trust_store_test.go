package qpid

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"net/http"
	"testing"
)

func TestAcceptanceTrustStore(t *testing.T) {

	_, certificateBytes, err := generateSelfSigned("Foo Org", "localhost")
	storeType := "NonJavaTrustStore"
	if err != nil {
		storeType = "SiteSpecificTrustStore"
	}
	certificateEncoded := certificateBytesToBase64(certificateBytes)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceTrustStoreCheckDestroy(testAcceptanceTrustStoreName),
		Steps: []resource.TestStep{
			{
				// test new trust store creation from configuration
				Config: getTrustStoreConfiguration(storeType, certificateEncoded),
				Check: testAcceptanceTrustStoreCheck(
					testAcceptanceTrustStoreResource,
					&map[string]interface{}{"name": testAcceptanceTrustStoreName},
				),
			},
			{
				// test trust store restoration from configuration after its deletion on broker side
				PreConfig: dropTrustStore(testAcceptanceTrustStoreName),
				Config:    getTrustStoreConfiguration(storeType, certificateEncoded),
				Check: testAcceptanceTrustStoreCheck(
					testAcceptanceTrustStoreResource,
					&map[string]interface{}{"name": testAcceptanceTrustStoreName},
				),
			},
			{
				// test trust store update
				Config: getTrustStoreConfigurationWithAttributes(storeType, certificateEncoded, `context = {"foo"="bar"}`),
				Check: testAcceptanceTrustStoreCheck(
					testAcceptanceTrustStoreResource,
					&map[string]interface{}{"name": testAcceptanceTrustStoreName,
						"context": map[string]interface{}{"foo": "bar"}},
				),
			},
			{
				// test trust store attribute removal
				Config: getTrustStoreConfiguration(storeType, certificateEncoded),
				Check: testAcceptanceTrustStoreCheck(
					testAcceptanceTrustStoreResource,
					&map[string]interface{}{"name": testAcceptanceTrustStoreName},
					"context",
				),
			},
		},
	})
}

func dropTrustStore(nodeName string) func() {
	return func() {
		client := testAcceptanceProvider.Meta().(*Client)
		resp, err := client.DeleteTrustStore(nodeName)
		if err != nil {
			fmt.Printf("unable to delete trust store: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			panic(fmt.Errorf("failed to delete trust store: %v", resp))
		}
	}
}

func testAcceptanceTrustStoreCheck(rn string, expectedAttributes *map[string]interface{}, removed ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s", rn)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("trust store id not set")
		}

		client := testAcceptanceProvider.Meta().(*Client)
		providers, err := client.GetTrustStores()
		if err != nil {
			return fmt.Errorf("error getting trust store: %s", err)
		}

		for _, provider := range *providers {
			if provider["id"] == rs.Primary.ID {
				return assertExpectedAndRemovedAttributes(&provider, expectedAttributes, removed)
			}
		}

		return fmt.Errorf("trust store '%s' is not found", rn)
	}
}

func testAcceptanceTrustStoreCheckDestroy(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAcceptanceProvider.Meta().(*Client)
		providers, err := client.GetTrustStores()
		if err != nil {
			return fmt.Errorf("error getting providers: %s", err)
		}

		for _, node := range *providers {
			if node["name"] == name {
				return fmt.Errorf("trust store '%v' still exists", node)
			}
		}

		return nil
	}
}

const testAcceptanceTrustStoreResourceName = "qpid_trust_store"
const testAcceptanceTrustStoreName = "acceptance_test_trust_store"
const testAcceptanceTrustStoreResource = testAcceptanceTrustStoreResourceName + "." + testAcceptanceTrustStoreName

func getTrustStoreConfiguration(storeType string, certificateEncoded string) string {
	return getTrustStoreConfigurationWithAttributes(storeType, certificateEncoded)
}

func getTrustStoreConfigurationWithAttributes(storeType string, certificateEncoded string, entries ...string) string {
	config := `
resource "` + testAcceptanceTrustStoreResourceName + `" "` + testAcceptanceTrustStoreName + `" {
    name = "` + testAcceptanceTrustStoreName + `"
    type = "` + storeType + `"
`
	if storeType == "NonJavaTrustStore" {
		config += `
    certificates_url = "data:;base64,` + certificateEncoded + `"
`
	} else {
		config += `
    site_url = "https://google.com"
`
	}

	for _, v := range entries {
		config += fmt.Sprintf("    %v\n", v)
	}
	config += `}
`

	return config
}
