package ink

import (
	"strings"
	"errors"
	"fmt"
	"reflect"
)

const (
	// InkClientAgent is the user agent header
	InkClientAgent  = "Ink-http-client/1.1"
	HTTP_USER_AGENT = "User-Agent"
)

func extractFromType(a interface{}) map[string]interface{} {
	var extracted = make(map[string]interface{})

	fields := reflect.TypeOf(a)
	values := reflect.ValueOf(a)

	num := fields.NumField()

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		value := values.Field(i)

		extracted[field.Name] = value
	}

	return extracted
}

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