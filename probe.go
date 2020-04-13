package rubik

import (
	"net/http/httptest"
	"time"
)

// GetTestClient returns a probed client for testing your rubik server
//
// Example:
// 		var client *rubik.Client
// 		func init() {
//			routers.import()
//			client = rubik.GetTestClient()
// }
//
// 		func TestSomething(t *testing.T) {
//			resp, err := client.Get("/")
//			** TEST YOUR `resp` **
//}
func GetTestClient() *Client {
	s := httptest.NewServer(app.mux)
	return NewClient(s.URL, time.Second*60)
}
