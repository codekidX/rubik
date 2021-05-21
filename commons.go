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
	// clientAgent is the user agent header
	clientAgent = "rubik-http-client/1.1"
	// HeaderUserAgent supplies user-agent key
	headerUserAgent = "User-Agent"
)

func valuesToMap(values url.Values) map[string]interface{} {
	var bodyMap = make(map[string]interface{})
	for k, v := range values {
		val, _ := strconv.Atoi(v[0])
		bodyMap[k] = val
	}
	return bodyMap
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
		message := fmt.Sprintf("RubikParamsSubstitutionError: ink was not able to substitute params because of params count mismatch - $ count: %d and params given: %d", dollarCount, len(reqParams))
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
	elem := reflect.TypeOf(entity)
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	_, ok := elem.FieldByName("Entity")
	return ok
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

func isOneOf(t string, vals ...string) bool {
	for _, s := range vals {
		if t == s {
			return true
		}
	}
	return false
}
