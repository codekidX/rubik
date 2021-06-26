package rubik

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/julienschmidt/httprouter"
)

// inject is the the entry point of request injection in rubik
// an injection is a process of reading the
func inject(req *http.Request,
	pm httprouter.Params, en interface{}, v Validation) (interface{}, error) {
	// lets check what type of request it is
	ctype := req.Header.Get(Content.Header)
	var body = make(map[string]interface{})
	var params = make(map[string]string)
	// check if any params in the route
	if len(pm) > 0 {
		for _, p := range pm {
			exportedKey := capitalize(p.Key)
			params[exportedKey] = p.Value
		}
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	switch ctype {
	case Content.JSON:
		err = json.Unmarshal(b, &body)
		if err != nil {
			return nil, err
		}
	case Content.URLEncoded:
		var encs url.Values
		encs, err = url.ParseQuery(string(b))
		// normalize the http.Values type to flat map
		for k, v := range encs {
			body[k] = v[0]
		}
	// TODO: need to evalueate multipart form data
	case Content.Multipart:
		err := req.ParseMultipartForm(32 << 20)
		if err != nil {
			return nil, err
		}
		break
	}

	values := reflect.ValueOf(en)
	fields := values.Elem().Type()
	num := values.Elem().NumField()

	for i := 0; i < num; i++ {
		field := fields.Field(i)

		if field.Name == "Entity" {
			continue
		}

		tag := field.Tag.Get(rubikTag)
		value := values.Elem().Field(i)
		transport := "query"
		transportKey := unCapitalize(field.Name)
		isRequired := false

		if strings.Contains(tag, "!") {
			isRequired = true
			tag = strings.ReplaceAll(tag, "!", "")
		}
		// get information from the tag
		if tag != "" {
			if strings.Contains(tag, "|") {
				reqTag := strings.Split(tag, "|")
				if reqTag[0] != "" {
					transportKey = reqTag[0]
				}
				if reqTag[1] != "" {
					transport = reqTag[1]
				}

			} else {
				if isOneOf(tag, "query", "body", "form", "param") {
					transport = tag
				} else {
					transportKey = tag
				}
			}
		}

		msg := "Data: %s is required but not found inside %s."
		requiredError := errors.New(fmt.Sprintf(msg, transportKey, transport))
		var val interface{}
		switch transport {
		case "query":
			val = req.URL.Query().Get(transportKey)
			if isRequired && (val == "" || val == nil) {
				return nil, requiredError
			}
			break
		case "body":
			val = body[transportKey]
			if (val == nil || val == "") && isRequired {
				return nil, requiredError
			}
			break
		case "form":
			val = req.Form.Get(transportKey)
			if (val == "") && isRequired {
				return nil, requiredError
			}
			break
		case "param":
			paramKey := capitalize(strings.ToLower(transportKey))
			val = params[paramKey]
			if val == "" && isRequired {
				return nil, requiredError
			}
			break
		}

		// this is for the validations the developer provided
		if len(v) > 0 && len(v[field.Name]) != 0 {
			for _, asrt := range v[field.Name] {
				err := asrt(val)
				if err != nil && strings.Contains(err.Error(), "$") {
					return nil, errors.New(strings.ReplaceAll(err.Error(), "$", field.Name))
				} else if err != nil {
					return nil, err
				}
			}
		}

		injectValueByType(val, value, field.Type.Kind())
	}

	return en, nil
}

func injectValueByType(val interface{}, elem reflect.Value, kind reflect.Kind) {
	switch kind {
	case reflect.String:
		final, ok := val.(string)
		if ok && elem.CanSet() {
			elem.SetString(final)
		}
		break
	case reflect.Int:
		value, _ := val.(string)
		if value == "" {
			return
		}

		final, ok := strconv.Atoi(value)
		if ok == nil && elem.CanSet() {
			elem.SetInt(int64(final))
		}
		break
	case reflect.Float32:
		value, _ := val.(string)
		if value == "" {
			return
		}
		// TODO: can make this conversion a function
		if floatVal, err := strconv.ParseFloat(value, 32); err == nil && elem.CanSet() {
			elem.SetFloat(floatVal)
		}
		break
	case reflect.Float64:
		value, _ := val.(string)
		if value == "" {
			return
		}
		// TODO: can make this conversion a function
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil && elem.CanSet() {
			elem.SetFloat(floatVal)
		}
		break
	case reflect.Bool:
		value, _ := val.(string)
		if value == "" {
			return
		}

		boolean, err := strconv.ParseBool(value)
		if err == nil && elem.CanSet() {
			elem.SetBool(boolean)
		}
		break
	case reflect.Struct:
		// should we loop a on all struct fields and add value?
		break
	case reflect.Slice:
		break
	}
}
