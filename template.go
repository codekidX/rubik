package rubik

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	html "html/template"
	text "text/template"

	"github.com/rubikorg/rubik/pkg"
)

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

// Render mehtod renders
func Render(path string, vars interface{}, ttype templateType) (RenderMixin, error) {
	// check for path of template folder
	templDir := pkg.GetTemplateFolderPath()
	templPath := templDir + string(os.PathSeparator) + strings.TrimPrefix(path, "/")
	if _, err := os.Stat(templPath); os.IsNotExist(err) {
		return RenderMixin{}, errors.New("templates directory not found")
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

	// check if the vars given by calling Parse function has some error
	return RenderMixin{contentType: contentType, content: content}, err
}

func parseTextTemplate(path, templName string, data interface{}) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t, err := text.New(templName).Parse(string(b))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)

	return buf.Bytes(), err
}

func parseHTMLTemplate(path, templName string, data interface{}) ([]byte, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	t, err := html.New(templName).Parse(string(b))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)

	return buf.Bytes(), err
}
