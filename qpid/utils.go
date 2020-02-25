package qpid

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func convertToCamelCase(name string) string {
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")

		result := ""
		for _, r := range parts[1:] {
			result += capitaliseString(r)
		}
		return parts[0] + result
	}
	return name
}

func convertToUnderscore(name string) string {
	result := ""
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			result += "_" + string(unicode.ToLower(r))
		} else {
			result += string(r)
		}
	}
	return result
}

func capitaliseString(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func createMapWithKeysInCameCase(m *map[string]interface{}) *map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range *m {
		newKey := convertToCamelCase(k)
		result[newKey] = v
	}
	return &result
}

func createMapWithKeysUnderscored(m *map[string]interface{}) *map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range *m {
		newKey := convertToUnderscore(k)
		result[newKey] = v
	}
	return &result
}

func convertToArrayOfStrings(items *[]interface{}) *[]string {
	var arr = make([]string, len(*items))
	for i, r := range *items {
		arr[i] = fmt.Sprintf("%v", r)
	}
	return &arr
}

func convertToMapOfStrings(m *map[string]interface{}) *map[string]string {
	result := make(map[string]string)
	for k, v := range *m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return &result
}

func convertIfValueIsStringWhenPrimitiveIsExpected(value interface{}, schemaType schema.ValueType) (interface{}, error) {
	var stringValue string
	var err error = nil
	var isString = false
	if value != nil {
		stringValue, isString = value.(string)
	}
	if isString {
		switch schemaType {
		case schema.TypeInt:
			log.Printf("Converting string '%s' to int", stringValue)
			value, err = strconv.Atoi(stringValue)
		case schema.TypeBool:
			log.Printf("Converting string '%s' to bool", stringValue)
			value, err = strconv.ParseBool(stringValue)
		case schema.TypeFloat:
			log.Printf("Converting string '%s' to float", stringValue)
			value, err = strconv.ParseFloat(stringValue, 64)
		}
	}
	return value, err
}

func convertHttpResponseToMap(res *http.Response) (*map[string]interface{}, error) {
	var err error
	defer func() {
		closeError := res.Body.Close()
		if err == nil {
			err = closeError
		}
	}()

	if res.StatusCode >= http.StatusNotFound {
		return &map[string]interface{}{}, nil
	}

	if res.StatusCode >= http.StatusBadRequest {
		return getErrorResponse(res)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &map[string]interface{}{}, err
	}

	// try decoding as map
	var m map[string]interface{}
	err = json.Unmarshal([]byte(body), &m)
	if err != nil {
		// try decoding as array of maps
		var arr []map[string]interface{}
		err2 := json.Unmarshal([]byte(body), &arr)
		if err2 == nil && len(arr) == 1 {
			m = arr[0]
		} else {
			return &map[string]interface{}{}, err
		}
	}

	return &m, nil
}

func convertHttpResponseToArray(res *http.Response) (*[]map[string]interface{}, error) {
	var err error
	defer func() {
		closeError := res.Body.Close()
		if err == nil {
			err = closeError
		}
	}()

	if res.StatusCode >= http.StatusNotFound {
		return &[]map[string]interface{}{}, nil
	}

	if res.StatusCode >= http.StatusBadRequest {
		_, err = getErrorResponse(res)
		return &[]map[string]interface{}{}, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &[]map[string]interface{}{}, err
	}

	// try decoding as array of maps
	var result []map[string]interface{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return &[]map[string]interface{}{}, err
	}

	return &result, nil
}

func getErrorResponse(res *http.Response) (*map[string]interface{}, error) {
	var rme map[string]interface{}
	err := json.NewDecoder(res.Body).Decode(&rme)
	if rme["message"] != nil {
		err = fmt.Errorf("error : %s", rme["message"])
	}
	return &rme, err
}

func schemaToAttributes(d *schema.ResourceData, schemaMap map[string]*schema.Schema, exclude ...string) *map[string]interface{} {
	attributes := make(map[string]interface{})
	excludes := arrayOfStringsToMap(exclude)
	for key := range schemaMap {
		if _, excluded := excludes[key]; excluded {
			continue
		}
		value, exists := d.GetOk(key)
		if exists {
			attributes[convertToCamelCase(key)] = value
		} else {
			oldValue, newValue := d.GetChange(key)
			if fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", newValue) {
				attributes[convertToCamelCase(key)] = nil
			}
		}
	}
	return &attributes
}

func arrayOfStringsToMap(slice []string) map[string]struct{} {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	return set
}

func containsExpectedAttributes(actual *map[string]interface{}, expected *map[string]interface{}) bool {
	for k, v := range *expected {
		if val, ok := (*actual)[k]; ok {
			if !reflect.DeepEqual(v, val) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func containsKeys(actual *map[string]interface{}, key []string) bool {
	for _, k := range key {
		if _, ok := (*actual)[k]; ok {
			return true
		}
	}
	return false
}

func assertExpectedAndRemovedAttributes(actual *map[string]interface{}, expectedAttributes *map[string]interface{}, removed []string) error {
	if expectedAttributes != nil && !containsExpectedAttributes(actual, expectedAttributes) {
		return fmt.Errorf("expected attributes are not found %v in %v", expectedAttributes, *actual)
	}
	if removed != nil && containsKeys(actual, removed) {
		return fmt.Errorf("one or more attributes from '%v' was not removed", removed)
	}

	return nil
}

func applyResourceAttributes(d *schema.ResourceData, attributes *map[string]interface{}, exclude ...string) error {
	if len(*attributes) == 0 {
		return nil
	}

	excludes := arrayOfStringsToMap(exclude)
	schemaMap := resourceKeyStore().Schema
	for key, v := range schemaMap {
		if _, excluded := excludes[key]; excluded {
			continue
		}

		_, keySet := d.GetOk(key)
		keyCamelCased := convertToCamelCase(key)
		value, attributeSet := (*attributes)[keyCamelCased]

		if keySet || attributeSet {
			value, err := convertIfValueIsStringWhenPrimitiveIsExpected(value, v.Type)
			if err != nil {
				return err
			}

			if v.Sensitive {
				continue
			}

			err = d.Set(key, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateSelfSigned(subject, host string) (*rsa.PrivateKey, *[]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %s", err)
		return nil, nil, err
	} else {
		notBefore := time.Now()
		notAfter := notBefore.Add(365 * 24 * time.Hour)
		serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
		serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
		if err != nil {
			log.Fatalf("Failed to generate serial number: %s", err)
			return nil, nil, err
		}

		template := x509.Certificate{
			SerialNumber: serialNumber,
			Subject: pkix.Name{
				Organization: []string{subject},
			},
			NotBefore:             notBefore,
			NotAfter:              notAfter,
			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: false,
			DNSNames:              []string{host},
		}

		certificateBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			log.Fatalf("Failed to create certificate: %s", err)
			return nil, nil, err
		}
		return privateKey, &certificateBytes, nil
	}
}

func certificateBytesToBase64(certificateBytes *[]byte) string {
	return base64.StdEncoding.EncodeToString(*certificateBytes)
}

func privateKeyToBase64(privateKey *rsa.PrivateKey) (string, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(privateKeyBytes), nil
}
