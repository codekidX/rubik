package rubik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	html "html/template"
	text "text/template"

	"github.com/rubikorg/rubik/pkg"
)

type TemplateStruct struct {
	Static string
}

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

type templateType int

// Type is a rubik type literal used for indication of response/template types
var Type = struct {
	HTML templateType
	JSON templateType
	Text templateType
}{1, 2, 3}

// Render returns a mixin holding the data to be rendered on the web page or
// sent over the wire
func Render(path string, vars interface{}, ttype templateType) (RenderMixin, error) {
	// check for path of template folder
	templDir := pkg.GetTemplateFolderPath()
	templPath := templDir + string(os.PathSeparator) + strings.TrimPrefix(path, "/")
	if _, err := os.Stat(templPath); os.IsNotExist(err) {
		msg := fmt.Sprintf("FileNotFoundError: path %s does not exist", templPath)
		return RenderMixin{}, errors.New(msg)
	}

	// check which templating to use html or text
	var contentType string
	var content []byte
	var err error
	switch ttype {
	case Type.HTML:
		contentType = "text/html"
		content, err = parseHTMLTemplate(templPath, "htmltempl", vars)
		break
	case Type.JSON, Type.Text:
		contentType = "text/plain"
		content, err = parseTextTemplate(templPath, "texttempl", vars)
		break
	}

	return RenderMixin{contentType: contentType, content: content}, err
}

// ParseDir returns a map of parsed template with key as file name and value as the
// content of the template of the given directory inside templates/ folder
//
// NOTE: this is not to be used as a controller response, this method only encapsulates
// logic around ParseFiles and makes it easy for developers to handle custom
// implementations
func ParseDir(dirName string, vars interface{}, ttype templateType) (map[string]string, error) {
	var result map[string]string
	folder := pkg.GetTemplateFolderPath() + string(os.PathSeparator) + dirName
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, f := range files {
		filePath := folder + string(os.PathSeparator) + f.Name()
		if ttype == Type.Text || ttype == Type.JSON {
			b, err := parseTextTemplate(filePath, f.Name(), vars)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			result[f.Name()] = string(b)
		} else {
			b, err := parseHTMLTemplate(filePath, f.Name(), vars)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			result[f.Name()] = string(b)
		}

	}

	return result, nil
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
