package rubik

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	// ClientAgent is the user agent header
	ClientAgent = "rubik-http-client/1.1"
	// HeaderUserAgent supplies user-agent key
	HeaderUserAgent = "User-Agent"
	// ContentType http header const
	ContentType = "Content-Type"
	// ContentJSON is json content type
	ContentJSON = "application/json"
	// ContentURLEncoded is url-encoded content type
	ContentURLEncoded = "application/x-www-form-urlencoded"
	// ContentMultipart is url-encoded content type
	ContentMultipart = "multipart/form-data"
)

func valuesToMap(values url.Values) map[string]interface{} {
	var bodyMap = make(map[string]interface{})
	for k, v := range values {
		val, _ := strconv.Atoi(v[0])
		bodyMap[k] = val
	}
	return bodyMap
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

func checkIsEntity(entity interface{}) bool {
	fields := reflect.TypeOf(entity)
	result := false
	num := fields.NumField()
	for i := 0; i < num; i++ {
		if fields.Field(i).Name == "RequestEntity" {
			return true
		}
	}
	return result
}

func isEmptyEntity(entity struct{}) bool {
	refEn := reflect.TypeOf(entity).Elem()
	newEn := reflect.New(refEn)
	return reflect.DeepEqual(entity, newEn)
}

func safeRouterPath(path string) string {
	if strings.HasSuffix(path, "/") {
		return strings.TrimSuffix(path, "/")
	}
	return path
}

func safeRoutePath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

func unCapitalize(target string) string {
	r := []rune(target)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func capitalize(target string) string {
	r := []rune(target)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
