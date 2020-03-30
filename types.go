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

// RequestEntity holds the data for a single API call
// It lets you write consolidated clean Go code
type RequestEntity struct {
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

// GetCtx ...
// func (re RequestEntity) GetCtx() *Payload {
// 	return re.request
// }

// BlankRequestEntity ...
type BlankRequestEntity struct {
	RequestEntity
}

// DownloadRequestEntity ...
type DownloadRequestEntity struct {
	RequestEntity
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
	Path     string
	OSFile   *os.File
	Metadata interface{}
}
