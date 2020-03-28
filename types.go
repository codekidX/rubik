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
	HTML ByteType
	JSON ByteType
	Text ByteType
}{1, 2, 3}

// ByteResponse is the response of rubik server
type ByteResponse struct {
	Data interface{}
	Type ByteType
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
