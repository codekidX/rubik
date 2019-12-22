package ink

import (
	"reflect"
)

const (
	// InkClientAgent is the user agent header
	InkClientAgent = "Ink-http-client/1.1"
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
