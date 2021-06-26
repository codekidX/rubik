package rubik

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	rubikTag = "rubik"
)

func getValueByKind(val reflect.Value, kind reflect.Kind) string {
	switch kind {
	case reflect.Bool:
		str := strconv.FormatBool(val.Bool())
		return str
	default:
		return val.String()
	}
}

func extract(en interface{}) (Payload, error) {
	var payload = Payload{}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fields := reflect.TypeOf(en)
	values := reflect.ValueOf(en)

	num := fields.NumField()

	for i := 0; i < num; i++ {
		field := fields.Field(i)
		tag := field.Tag.Get("rubik")
		value := values.Field(i)

		// Get route value from Entity
		if (field.Type == reflect.TypeOf(Entity{})) {
			route := value.FieldByName("PointTo").String()
			params, ok := value.FieldByName("Params").Interface().([]string)
			if ok {
				payload.params = params
			}
			isJSON := value.FieldByName("JSON").Bool()
			isURLEncoded := value.FieldByName("URLEncoded").Bool()
			isFormData := value.FieldByName("FormData").Bool()

			infer := value.FieldByName("Infer").Interface()

			if infer != nil {
				payload.responseType = infer
			}

			payload.json = isJSON
			payload.urlencoded = isURLEncoded
			payload.formData = isFormData
			payload.path = route
			if route == "" {
				return Payload{}, errors.New("ExtractionError: Entity is initialized " +
					"without a Route parameter")
			}
			continue
		}

		var op = strings.ReplaceAll(tag, "?", "")
		op = strings.ReplaceAll(tag, "*", "")
		var transportKey = unCapitalize(field.Name)
		var transport = "query"
		if strings.Contains(op, "|") {
			reqTag := strings.Split(op, "|")
			// one more check for checking if pipe operator is used then check if
			// it is only used for transport specification
			if reqTag[1] == "" {
				return Payload{}, errors.New("MalformedTag: rubik tags must be in the form of " +
					"[optional:key]|[transport] found: " + tag)
			}
			transportKey = reqTag[0]
			transport = reqTag[1]
		}

		switch transport {
		case "param":
			continue
		case "body":
			if payload.body == nil {
				payload.body = Values{}
			}

			exVal := values.Field(i).Elem().Interface()
			strVal, ok := exVal.(string)
			if !ok {
				payload.body.Set(transportKey, exVal)
			} else if strVal != "" {
				payload.body.Set(transportKey, strVal)
			}
			break
		case "query":
			if payload.query == nil {
				payload.query = url.Values{}
			}
			val := getValueByKind(value, value.Kind())
			payload.query.Set(transportKey, val)
			break
		case "form":
			if field.Type == reflect.TypeOf(File{}) {
				err := extractFileInfo(payload, values.Field(i).Elem().Interface(), transportKey,
					writer)
				if err != nil {
					return Payload{}, err
				}
			} else if field.Type == reflect.TypeOf([]File{}) {
				files, _ := values.Field(i).Interface().([]File)

				for _, f := range files {
					err := extractFileInfo(payload, f, transportKey, writer)
					if err != nil {
						return Payload{}, err
					}
				}

			} else {
				writer.WriteField(transportKey, value.String())
			}
			break
		default:
			message := fmt.Sprintf("UnknownMedium: Not a medium: %s", transport)
			return Payload{}, errors.New(message)
		}

	}

	if body.Len() > 0 && payload.formData {
		payload.headers = url.Values{}
		payload.headers.Set("Content-Type", writer.FormDataContentType())
	}

	err := writer.Close()
	if err != nil {
		return Payload{}, err
	}

	payload.formBody = body

	return payload, nil
}

func extractFileInfo(
	payload Payload, value interface{}, key string, writer *multipart.Writer) error {
	f, _ := value.(File)
	if f.Path != "" {
		file, err := os.Open(f.Path)

		if err != nil {
			return err
		}

		defer file.Close()

		part, err := writer.CreateFormFile(key, file.Name())

		if err != nil {
			return err
		}

		_, err = io.Copy(part, file)

		if err != nil {
			return err
		}
	} else if f.OSFile.Name() != "" {
		defer f.OSFile.Close()

		part, err := writer.CreateFormFile(key, f.OSFile.Name())

		if err != nil {
			return err
		}

		_, err = io.Copy(part, f.OSFile)

		if err != nil {
			return err
		}
	}
	return nil
}
