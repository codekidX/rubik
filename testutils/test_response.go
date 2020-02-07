package testutils

import (
	"net/http"
)

// TestSimpleGetResponse ...
type TestSimpleGetResponse struct {
	Greetings string `json:"greetings"`
}

// WriteJSON writes JSON bytes to given response writer
func WriteJSON(data []byte, rw http.ResponseWriter) {

}