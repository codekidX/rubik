package rubik

import (
	"net/http/httptest"
	"time"
)

// GetTestClient returns a probed client for testing
func GetTestClient() *Client {
	s := httptest.NewServer(app.mux)
	return NewClient(s.URL, time.Second*60)
}
