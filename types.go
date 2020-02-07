package cherry

import (
	"fmt"
	"os"
)

// Handler ...
type Handler struct {
	Entity RequestEntity
}

// RequestEntity holds the data for a single API call
// It lets you write consolidated clean Go code
type RequestEntity struct {
	entityType string
	route      string
	Params     []string
	request    *Request
	FormData   bool
	URLEncoded bool
	JSON       bool
	Infer      interface{}
}

// Route ...
func (re *RequestEntity) Route(path string, values ...string) {
	re.route = path
	re.Params = values
}

// GetCtx ...
func (re RequestEntity) GetCtx() *Request {
	return re.request
}

type BlankRequestEntity struct {
	RequestEntity
}

// Values that transend url.Values with strict typing
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
