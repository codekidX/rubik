package rubik

import (
	"reflect"
	"testing"
)

var probe TestProbe

func init() {
	probe = NewProbe(initTestRouter())
}

func initTestRouter() Router {
	indexRouter := Create("/")
	i := Route{
		Path:       "/",
		Controller: testIndexCtl,
	}
	indexRouter.Add(i)
	return indexRouter
}

func testIndexCtl(req *Request) {
	req.Respond("Woohoo!")
}

func TestGetTestClient(t *testing.T) {
	if reflect.TypeOf(probe) != reflect.TypeOf(TestProbe{}) {
		t.Error("Probe did not return a value of type rubik.TestProbe")
	}
}

func TestGetCallWithTestClient(t *testing.T) {
	_, rr := probe.Test("GET", "/", nil, nil, testIndexCtl)

	if rr.Result().StatusCode != 200 {
		t.Error("Router for index initialized but request returned non 200 response code:",
			rr.Result().StatusCode)
	}
}
