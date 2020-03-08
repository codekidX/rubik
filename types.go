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
	request    *Payload
	FormData   bool
	URLEncoded bool
	JSON       bool
	Infer      interface{}
}

// Route ...
func (re RequestEntity) Route(path string, values ...string) RequestEntity {
	re.route = path
	re.Params = values
	return re
}

// Route ...
func (re DownloadRequestEntity) Route(path string, values ...string) DownloadRequestEntity {
	re.route = path
	re.Params = values
	return re
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
