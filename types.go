package rubik

import (
	"fmt"
	"net/http"
	"os"
)

// ByteType let's rubik know what type of bytes to send as response
type ByteType int

// Type is a rubik type literal used for indication of response/template types
var Type = struct {
	HTML         ByteType
	JSON         ByteType
	Text         ByteType
	Bytes        ByteType
	templateHTML ByteType
	templateText ByteType
}{1, 2, 3, 4, 5, 6}

// Content is a struct that holds default values of Content-Type headers
// it can be used throughout your rubik application for avoiding basic
// spell mistakes
var Content = struct {
	Header     string
	JSON       string
	Text       string
	HTML       string
	URLEncoded string
	Multipart  string
}{
	"Content-Type",
	"application/json",
	"text/plain",
	"text/html",
	"application/x-www-form-urlencoded",
	"multipart/form-data",
}

var StringByteTypeMap = map[string]ByteType{
	"json": Type.JSON,
	"html": Type.HTML,
	"text": Type.Text,
}

// Message that is to be sent in communicator channel
type Message struct {
	Communicator string
	Topic        string
	Body         interface{}
}

// MessagePasser holds the channels for communication with rubik server
type MessagePasser struct {
	Message chan Message
	Error   chan error
}

// ByteResponse is the response of rubik server
type ByteResponse struct {
	Status      int
	Data        interface{}
	OfType      ByteType
	Error       error
	redirectURL string
	safeWrite   bool
	handler     http.Handler
	handlerFunc http.HandlerFunc
}

// Validation is validation operations to be performed
// on the request entity
type Validation map[string][]Assertion

// SessionManager is an interface contract that rubik.Session uses
// Anything abiding by this contract can
type SessionManager interface {
	Get(string) string
	Set(string, string) error
	Delete(string) bool
}

// Communicator interface is used to handle the service/driver that
// rubik's inherent communication depends upon
type Communicator interface {
	Send(string, interface{}) error
}

// Entity holds the data for a single API call
// It lets you write consolidated clean Go code
type Entity struct {
	entityType string
	PointTo    string
	Params     []string
	request    *http.Request
	FormData   bool
	URLEncoded bool
	JSON       bool
	Infer      interface{}
	Cookies    Values
}

// GobEntity extends Entity and let's rubik know that this date is
// present inside body of the request and you need to decode it using
// GobDecoder
// type GobEntity struct {
// 	Entity
// }

// GetCtx ...
// func (re RequestEntity) GetCtx() *Payload {
// 	return re.request
// }

// BlankRequestEntity ...
type BlankRequestEntity struct {
	Entity
}

// DownloadRequestEntity ...
type DownloadRequestEntity struct {
	Entity
	TargetFilePath string
}

// Values that transcends url.Values allowing any type as value
type Values map[string]interface{}

// Encode converts the values into urlencoded strings
func (val Values) Encode() string {
	var qs = ""
	var cnt = 0
	if len(val) > 0 {
		for k, v := range val {
			qs += fmt.Sprintf("%s=%v", k, v)
			if cnt+1 != len(val) {
				qs += "&"
			}
			cnt++
		}
	}
	return qs
}

// Set assigns a value for key `key` inside the rubik.Values
func (val Values) Set(key string, value interface{}) {
	val[key] = value
}

// File used by ink to embbed file
type File struct {
	Path   string
	OSFile *os.File
	Raw    []byte
}

// RResponseWriter is Rubik's response writer that implements http.ResponseWriter
// and it's methods to provide additional functionalities related to Rubik.
type RResponseWriter struct {
	http.ResponseWriter
	written bool
	status  int
	data    []byte
}

// WriteHeader writes the http.Request's Header values to the wire and
// sets the status given as the parameter
func (w *RResponseWriter) WriteHeader(status int) {
	w.status = status
	w.written = true
	w.ResponseWriter.WriteHeader(status)
}

// Write writes the response bytes b to the wire
func (w *RResponseWriter) Write(b []byte) (int, error) {
	w.data = b
	w.written = true
	return w.ResponseWriter.Write(b)
}

// Assertion is the assert functions for rubiks validation cycle
// it should return a message stating why validation failed and
// bool indicating if assertion has passed or not
type Assertion func(interface{}) error

// Modifier lets you modify the parameters passed through the request.
// The input of the modifier is the value from the wire and the output
// is the modified value.
// Note: Modified value type must match the type declared in the
// entity.
// type Modifier func(interface{}) interface{}
