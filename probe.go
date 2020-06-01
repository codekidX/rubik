package rubik

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
)

// TestProbe is an abstraction for easily testing your rubik routes
type TestProbe struct {
	app *rubik
}

// NewProbe returns a probe for testing your rubik server
//
// Example:
// 		var probe rubik.TestProbe
// 		func init() {
//			// pass the rubik.Router you want to test
//			probe = rubik.NewProbe(index.Router)
// 		}
//
// 		func TestSomeRoute(t *testing.T) {
//			// returns the *http.Request, *httptest.ResponseRecorder used inside the test
//			req, rr := probe.Test(en)
//			if rr.Result().StatusCode != 200 { /* Something is wrong */}
//		}
func NewProbe(ro Router) TestProbe {
	os.Setenv("RUBIK_ENV", "test")
	// boot only inits the routes of the rubik server
	// without inititializing the app or running the
	// server
	Use(ro)
	boot(false)
	p := TestProbe{}
	p.app = app
	return p
}

// Test a route with method, path to request, Entity (if used) and the controller to test
func (p TestProbe) Test(method, path string, reqBody io.Reader, en interface{},
	ctl Controller) (*http.Request, *httptest.ResponseRecorder) {

	req := httptest.NewRequest(method, path, reqBody)
	rr := httptest.NewRecorder()
	rubikReq := Request{
		Entity: en,
		Raw:    req,
		Writer: RResponseWriter{ResponseWriter: rr},
	}

	ctl(&rubikReq)
	return req, rr
}

// TestHandler is a probe util function to test your handler if you are not using a
// rubik.Controller for your route and using UseHandler() to cast it
func (p TestProbe) TestHandler(method, path string, reqBody io.Reader, en interface{},
	h http.Handler) (*http.Request, *httptest.ResponseRecorder) {

	req := httptest.NewRequest(method, path, reqBody)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)
	return req, rr

}
