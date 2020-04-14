package rubik

import (
	"net/http/httptest"
	"time"
)

// Probe returns a probed client for testing your rubik server
//
// Example:
// 		var client *rubik.Client
// 		func init() {
//			routers.import()
//			client = rubik.Probe()
// }
//
// 		func TestSomething(t *testing.T) {
//			resp, err := client.Get("/")
//			** TEST YOUR `resp` **
//}
func Probe() *Client {
	s := httptest.NewServer(app.mux)
	boot(false)
	return NewClient(s.URL, time.Second*60)
}
