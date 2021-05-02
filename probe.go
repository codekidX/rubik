package rubik

import (
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/rubikorg/rubik/pkg"
)

// TestableEntity is an entity which can be probed by the Rubik probe
// framework
type TestableEntity interface {
	ComposedEntity() Entity
	CoreEntity() interface{}
	Path() string
}

// TestProbe is an abstraction for easily testing your rubik routes
type TestProbe struct {
	app    *rubik
	router Router
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
func NewProbe(ro Router) *TestProbe {
	os.Setenv("RUBIK_ENV", "test")
	var a = make(map[string]interface{})
	// use the router in this package
	Use(ro)
	// load some default dummy config
	Load(&a)
	boot(false, false)
	p := TestProbe{}
	p.app = app
	p.router = ro
	return &p
}

// Test will test your entity with given `testPath` on the given Rubik Router
// to the probe using the rubik.NewProbe() func
func (probe *TestProbe) Test(entity TestableEntity) *httptest.ResponseRecorder {
	return probe.fetchResponse(entity)
}

// TestAll performs the same operation as Test but performs it for given slice of
// entities
func (probe *TestProbe) TestAll(entities []TestableEntity) []*httptest.ResponseRecorder {
	var allResponses []*httptest.ResponseRecorder
	for _, e := range entities {
		allResponses = append(allResponses, probe.fetchResponse(e))
	}
	return allResponses
}

func (probe *TestProbe) fetchResponse(entity TestableEntity) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	if entity == nil {
		return rr
	}

	r := probe.getRouteFromEntity(entity)
	pathSuffix := safeRoutePath(entity.Path())
	if entity.Path() == "" {
		pkg.DebugMsg("Test | entity.PointTo is empty, using / as the endpoint locator")
		pathSuffix = "/"
	}

	finalPath := probe.app.url + safeRouterPath(probe.router.basePath) + pathSuffix

	req, _ := http.NewRequest(getSafeMethod(r.Method), finalPath, nil)
	rubikReq := Request{
		Entity: entity,
		Raw:    req,
		Writer: RResponseWriter{ResponseWriter: rr},
	}

	r.Controller(&rubikReq)
	return rr
}

func (probe *TestProbe) getRouteFromEntity(entity TestableEntity) Route {
	for _, r := range probe.router.routes {
		if entity.Path() == r.Path {
			return r
		}
	}
	return Route{}
}

func getSafeMethod(method string) string {
	if method == "" {
		return http.MethodGet
	}
	return method
}
