package rubik

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	html "html/template"
	text "text/template"
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

// stackTraceTemplate is the data binded to errors.html templated rendered by
// rubik server
type stackTraceTemplate struct {
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
func Render(btype ByteType, vars interface{}, paths ...string) Controller {
	return func(req *Request) {
		bresp := RenderContent(btype, vars, paths...)
		if bresp.Error != nil {
			req.Throw(bresp.Status, bresp.Error)
			return
		}

		var ty string
		switch bresp.OfType {
		case Type.HTML:
			ty = Content.HTML
			break
		case Type.Text:
			ty = Content.Text
			break
		case Type.JSON:
			ty = Content.JSON
			break
		}

		writeResponse(&req.Writer, bresp.Status, ty, bresp.Data.([]byte))
	}
}

// RenderContent returns you the response bytes of the target paths that is going to be
// written on the wire. It is an abstraction of Render method which is used as
// a layered method inside Render method.
func RenderContent(btype ByteType, vars interface{}, paths ...string) ByteResponse {
	allPaths := []string{}
	var templPath string
	var multiple = false
	if len(paths) == 0 {
		return ByteResponse{Status: 500, Error: errors.New("No paths passed to Render method")}
	} else if len(paths) > 1 {
		for _, path := range paths {
			allPaths = append(allPaths, filepath.Join("templates", path))
		}
		multiple = true
	} else {
		templPath = filepath.Join("templates", paths[0])
	}

	var content []byte
	resType := Type.templateHTML
	var err error
	switch btype {
	case Type.HTML:
		if multiple {
			content, err = parseMultipleHTMLTemplate(allPaths, "htmltempl", vars)
		} else {
			content, err = parseHTMLTemplate(templPath, "htmltempl", vars)
		}
		break
	case Type.JSON, Type.Text:
		resType = Type.templateText
		if multiple {
			content, err = parseMultipleTextTemplate(allPaths, "texttempl", vars)
		} else {
			content, err = parseTextTemplate(templPath, "texttempl", vars)
		}
		break
	}

	if err != nil {
		return ByteResponse{Status: 500, Error: err}
	}

	return ByteResponse{Status: 200, Data: content, OfType: resType}
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
