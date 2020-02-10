package qpid

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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
