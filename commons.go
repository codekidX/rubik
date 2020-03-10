package sketch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	// ClientAgent is the user agent header
	ClientAgent = "cherry-http-client/1.1"
	// HeaderUserAgent supplies user-agent key
	HeaderUserAgent = "User-Agent"
)

// extractFromType takes in an RequestEntity and returns a final payload
// that is to be passed to the net/http module
func extractFromType(a interface{}) (Payload, error) {
	var extracted = Payload{}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fields := reflect.TypeOf(a)
	values := reflect.ValueOf(a)

	num := fields.NumField()

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		tag := field.Tag.Get("ink")
		value := values.Field(i)

		// Get route value from RequestEntity
		if (field.Type == reflect.TypeOf(RequestEntity{})) {
			route := value.FieldByName("route").String()
			params, ok := value.FieldByName("Params").Interface().([]string)
			if ok {
				extracted.params = params
			}
			isJSON := value.FieldByName("JSON").Bool()
			isURLEncoded := value.FieldByName("URLEncoded").Bool()
			isFormData := value.FieldByName("FormData").Bool()

			infer := value.FieldByName("Infer").Interface()

			if infer != nil {
				extracted.responseType = infer
			}

			extracted.json = isJSON
			extracted.urlencoded = isURLEncoded
			extracted.formData = isFormData
			extracted.path = route
			if route == "" {
				return Payload{}, errors.New("InkExtractionError: Ink RequestEntity is initialized without a Route parameter")
			}
			continue
		}

		var op = tag

		if strings.Index(tag, ",") > 0 {
			splitted := strings.Split(tag, ",")
			op = splitted[0]
		}

		if !strings.Contains(op, "|") {
			return Payload{}, errors.New("InkTagMalformed: Ink tag must be of format key|request_medium('body,query etc..') ")
		}

		reqTag := strings.Split(op, "|")

		switch reqTag[1] {
		// TODO: set by type of value
		case "body":
			if extracted.body == nil {
				extracted.body = Values{}
			}

			exVal := values.Field(i).Interface()
			strVal, ok := exVal.(string)
			if !ok {
				extracted.body.Set(reqTag[0], exVal)
			} else if strVal != "" {
				extracted.body.Set(reqTag[0], strVal)
			}
			break
		case "query":
			if extracted.query == nil {
				extracted.query = url.Values{}
				extracted.query.Set(reqTag[0], value.String())
			} else {
				extracted.query.Set(reqTag[0], value.String())
			}
			break
		case "form":
			if field.Type == reflect.TypeOf(File{}) {
				f, _ := values.Field(i).Interface().(File)
				if f.Path != "" {
					file, err := os.Open(f.Path)

					if err != nil {
						return Payload{}, err
					}

					defer file.Close()

					part, err := writer.CreateFormFile(reqTag[0], file.Name())

					if err != nil {
						return Payload{}, err
					}

					_, err = io.Copy(part, file)

					if err != nil {
						return Payload{}, err
					}
				} else if f.OSFile.Name() != "" {
					defer f.OSFile.Close()

					part, err := writer.CreateFormFile(reqTag[0], f.OSFile.Name())

					if err != nil {
						return Payload{}, err
					}

					_, err = io.Copy(part, f.OSFile)

					if err != nil {
						return Payload{}, err
					}
				}
			} else if field.Type == reflect.TypeOf([]File{}) {
				files, _ := values.Field(i).Interface().([]File)

				for _, f := range files {
					if f.Path != "" {
						file, err := os.Open(f.Path)

						if err != nil {
							return Payload{}, err
						}

						defer file.Close()

						part, err := writer.CreateFormFile(reqTag[0], file.Name())

						if err != nil {
							return Payload{}, err
						}

						_, err = io.Copy(part, file)

						if err != nil {
							return Payload{}, err
						}
					} else if f.OSFile.Name() != "" {
						defer f.OSFile.Close()

						part, err := writer.CreateFormFile(reqTag[0], f.OSFile.Name())

						if err != nil {
							return Payload{}, err
						}

						_, err = io.Copy(part, f.OSFile)

						if err != nil {
							return Payload{}, err
						}
					}
				}

			} else {
				writer.WriteField(reqTag[0], value.String())
			}
			break
		default:
			message := fmt.Sprintf("InkUnknownMedium: Not a medium: %s", reqTag[1])
			return Payload{}, errors.New(message)
		}

	}

	if body.Len() > 0 && extracted.formData {
		extracted.headers = url.Values{}
		extracted.headers.Set("Content-Type", writer.FormDataContentType())
	}

	err := writer.Close()

	if err != nil {
		return Payload{}, err
	}

	extracted.formBody = body

	return extracted, nil
}

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

// func checkReqInputs(entity interface{}) bool {
// 	fields := reflect.TypeOf(entity)
// 	values := reflect.ValueOf(entity)

// 	result := false
// 	num := fields.NumField()
// 	for i := 0; i < num; i++ {
// 		field := fields.Field(i)
// 		if (strings.Contains(field.Tag.Get("ink"), "omitempty") != "" &&
// 		values.Field(i).String() ==
// 		) {

// 		}
// 	}
// }
