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

type rx struct {
	ctl Controller
	en  interface{}
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
	Status int
	Data   interface{}
	OfType ByteType
	Error  error
}

// Validation is validation operations to be performed
// on the request entity
type Validation map[string]string

// SessionManager is an interface contract that rubik.Session uses
//
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

// AuthorizationGuard defines the requirement of a rubik route guard
// Guards are generally used to negate the request, eg: JWT, ACL etc..
// The implementation for a guard should be transparent with specifiction
// of realm and config requirement.
// For more information take a look at the BasicGuard implementation
// here: https://github.com/rubikorg/blocks/blob/master/guard/basic.go
type AuthorizationGuard interface {
	// Require specifies the config requirement of your Guard
	// this method must return the name of the config that is
	// to be checked before setting the WWW-Authenticate header
	Require() string
	// Require must return the value of the realm to be set
	// inside WWW-Authenticate header header
	GetRealm() string
	// Authorize holds the main authorization logic for a given guard
	// NOTE: Rubik does not proceed with the request if this
	// method returns an error. The error will denote HTTP Status: 401
	Authorize(*App, http.Header) error
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
type GobEntity struct {
	Entity
}

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

// Values that transcend url.Values with strict typing
type Values map[string]interface{}

// Encode ...
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

// Set ...
func (val Values) Set(key string, value interface{}) {
	val[key] = value
}

// File used by ink to embbed file
type File struct {
	Path   string
	OSFile *os.File
	Raw    []byte
}
