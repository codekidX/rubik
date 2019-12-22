package ink

import (
	"net/url"
	"strings"
	"errors"
	"fmt"
	"reflect"
)

const (
	// InkClientAgent is the user agent header
	InkClientAgent  = "Ink-http-client/1.1"
	// HeaderUserAgent supplies user-agent key 
	HeaderUserAgent = "User-Agent"
)

func extractFromType(a interface{}) url.Values {
	var extracted = url.Values{}

	fields := reflect.TypeOf(a)
	values := reflect.ValueOf(a)

	num := fields.NumField()

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		value := values.Field(i)

		extracted.Set(field.Name, value.String())
	}

	return extracted
}

// func injectToType(respMap map[string]interface{}, infer interface{}, target *interface{}) interface{} {
// 	fields := reflect.TypeOf(infer)

// 	num := fields.NumField()

// 	for i := 0; i < num; i++ {
// 		fields := fields.Field(i)

		
// 	}
// }

func getCountOfDollar(path string) int {
	count := 0
	for _, s := range path {
		if s == '$' {
			count++
		}
	}
	return count
}

func substituteParam(path string, reqParams []string) (string, error) {
	pathWithParams := path
	dollarCount := getCountOfDollar(pathWithParams)
	if len(reqParams) != dollarCount {
		message := fmt.Sprintf("InkParamsSubstitutionError: ink was not able to substitute params because of params count mismatch - $ count: %d and params given: %d", dollarCount, len(reqParams))
		return "", errors.New(message)
	}
	// we need a replaced path
	if len(reqParams) > 0 {
		for _, param := range reqParams {
			pathWithParams = strings.Replace(pathWithParams, "$", param, -1)
		}
	}

	return pathWithParams, nil
}