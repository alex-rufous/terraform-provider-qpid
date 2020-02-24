package qpid

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"log"
	"math/big"
	"net/http"
	"testing"
	"time"
)

func TestAcceptanceTrustStore(t *testing.T) {

	var certificateEncoded, privateKeyEncoded string
	storeType := "NonJavaTrustStore"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %s", err)
		storeType = "SiteSpecificTrustStore"
	} else {
		notBefore := time.Now()
		notAfter := notBefore.Add(365 * 24 * time.Hour)
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			log.Fatalf("Failed to generate serial number: %s", err)
			storeType = "SiteSpecificTrustStore"
		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{"Foo Org"},
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: false,
			DNSNames:              []string{"localhost"},
		}

		certificateBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			log.Fatalf("Failed to create certificate: %s", err)
			storeType = "SiteSpecificTrustStore"
		}

		certificateEncoded = base64.StdEncoding.EncodeToString(certificateBytes)
		privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			log.Fatalf("Unable to marshal private key: %v", err)
			storeType = "SiteSpecificTrustStore"
		}
		privateKeyEncoded = base64.StdEncoding.EncodeToString(privateKeyBytes)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAcceptancePreCheck(t) },
		Providers:    testAcceptanceProviders,
		CheckDestroy: testAcceptanceTrustStoreCheckDestroy(testAcceptanceTrustStoreName),
		Steps: []resource.TestStep{
			{
				// test new trust store creation from configuration
				Config: getTrustStoreConfiguration(storeType, privateKeyEncoded, certificateEncoded),
				Check: testAcceptanceTrustStoreCheck(
					testAcceptanceTrustStoreResource,
					&map[string]interface{}{"name": testAcceptanceTrustStoreName},
				),
			},
			{
				// test trust store restoration from configuration after its deletion on broker side
				PreConfig: dropTrustStore(testAcceptanceTrustStoreName),
				Config:    getTrustStoreConfiguration(storeType, privateKeyEncoded, certificateEncoded),
				Check: testAcceptanceTrustStoreCheck(
					testAcceptanceTrustStoreResource,
					&map[string]interface{}{"name": testAcceptanceTrustStoreName},
				),
			},
			{
				// test trust store update
				Config: getTrustStoreConfigurationWithAttributes(storeType, privateKeyEncoded, certificateEncoded, `context = {"foo"="bar"}`),
				Check: testAcceptanceTrustStoreCheck(
					testAcceptanceTrustStoreResource,
					&map[string]interface{}{"name": testAcceptanceTrustStoreName,
						"context": map[string]interface{}{"foo": "bar"}},
				),
			},
			{
				// test trust store attribute removal
				Config: getTrustStoreConfiguration(storeType, privateKeyEncoded, certificateEncoded),
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

func getTrustStoreConfiguration(storeType string, privateKeyEncoded string, certificateEncoded string) string {
	return getTrustStoreConfigurationWithAttributes(storeType, privateKeyEncoded, certificateEncoded)
}

func getTrustStoreConfigurationWithAttributes(storeType string, privateKeyEncoded string, certificateEncoded string, entries ...string) string {
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
