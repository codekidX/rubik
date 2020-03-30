package rubik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"

	html "html/template"
	text "text/template"

	"github.com/rubikorg/rubik/pkg"
)

const (
	sep = string(os.PathSeparator)
)

// TemplateStruct when extended provides basic integrations with the rubik server
// type TemplateStruct struct {
// 	Static string
// }

// // SocketURL ensures that your client has a proper transport for communiction after
// // render is done
// func (ts TemplateStruct) SocketURL() string {
// 	return app.url
// }

// StackTraceTemplate is the data binded to errors.html templated rendered by
// rubik server
type StackTraceTemplate struct {
	Msg   string
	Stack []string
}

// RenderMixin is a mixin holding values for rendering a template
type RenderMixin struct {
	content     []byte
	contentType string
}

// Result returns the parsed/executed content of a template as bytes
func (rm RenderMixin) Result() []byte {
	return rm.content
}

// Render returns a mixin holding the data to be rendered on the web page or
// sent over the wire
func Render(btype ByteType, vars interface{}, paths ...string) ByteResponse {
	// check for path of template folder
	templDir := pkg.GetTemplateFolderPath()
	absPaths := []string{}
	var templPath string
	var multiple = false
	if len(paths) == 0 {
		return ByteResponse{
			Status: http.StatusInternalServerError,
			Error:  errors.New("No paths passed to Render method"),
		}
	} else if len(paths) > 1 {
		for _, path := range paths {
			templPath := templDir + sep + strings.TrimPrefix(path, "/")
			if _, err := os.Stat(templPath); os.IsNotExist(err) {
				msg := fmt.Sprintf("FileNotFoundError: path %s does not exist", templPath)
				return ByteResponse{
					Status: http.StatusInternalServerError,
					Error:  errors.New(msg),
				}
			}
			absPaths = append(absPaths, path)
		}
		multiple = true
	} else {
		templPath = templDir + sep + strings.TrimPrefix(paths[0], "/")
	}

	var content []byte
	resType := Type.templateHTML
	var err error
	switch btype {
	case Type.HTML:
		if multiple {
			content, err = parseMultipleHTMLTemplate(absPaths, "htmltempl", vars)
		} else {
			content, err = parseHTMLTemplate(templPath, "htmltempl", vars)
		}
		break
	case Type.JSON, Type.Text:
		resType = Type.templateText
		if multiple {
			content, err = parseMultipleTextTemplate(absPaths, "texttempl", vars)
		} else {
			content, err = parseTextTemplate(templPath, "texttempl", vars)
		}
		break
	}

	if err != nil {
		return ByteResponse{
			Status: http.StatusInternalServerError,
			Error:  err,
		}
	} else {
		return ByteResponse{Status: 200, Data: content, OfType: resType}
	}
}

// ParseDir returns a map of parsed template with key as file name and value as the
// content of the template of the given directory inside templates/ folder
//
// NOTE: this is not to be used as a controller response, this method only encapsulates
// logic around ParseFiles and makes it easy for developers to handle custom
// implementations
func ParseDir(dirName string, vars interface{}, btype ByteType) ByteResponse {
	var result = make(map[string]string)
	folder := pkg.GetTemplateFolderPath() + string(os.PathSeparator) + dirName
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return Failure(500, errors.WithStack(err))
	}

	for _, f := range files {
		filePath := folder + string(os.PathSeparator) + f.Name()
		if btype == Type.Text || btype == Type.JSON {
			b, err := parseTextTemplate(filePath, f.Name(), vars)
			if err != nil {
				return Failure(500, errors.WithStack(err))
			}
			result[f.Name()] = string(b)
		} else {
			b, err := parseHTMLTemplate(filePath, f.Name(), vars)
			if err != nil {
				return Failure(500, errors.WithStack(err))
			}
			result[f.Name()] = string(b)
		}

	}

	return Success(result, Type.JSON)
}

func parseMultipleTextTemplate(
	paths []string, templName string, data interface{}) ([]byte, error) {
	t, err := text.New(templName).ParseFiles(paths...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)

	return buf.Bytes(), errors.WithStack(err)
}

func parseTextTemplate(path, templName string, data interface{}) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	t, err := text.New(templName).Parse(string(b))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// check if the vars given by calling Parse function has some error
	var buf bytes.Buffer
	err = t.Execute(&buf, data)

	return buf.Bytes(), errors.WithStack(err)
}

func parseMultipleHTMLTemplate(
	paths []string, templName string, data interface{}) ([]byte, error) {
	t, err := html.New(templName).ParseFiles(paths...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)

	return buf.Bytes(), errors.WithStack(err)
}

func parseHTMLTemplate(path, templName string, data interface{}) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	t, err := html.New(templName).Parse(string(b))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)

	return buf.Bytes(), errors.WithStack(err)
}
