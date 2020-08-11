package rubik

import (
	"reflect"
	"testing"
)

var probe TestProbe

func init() {
	probe = NewProbe(initTestRouter())
}

var ir = Route{
	Method:     "GET",
	Path:       "/",
	Controller: testIndexCtl,
}

func initTestRouter() Router {
	indexRouter := Create("/")
	indexRouter.Add(ir)
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

func TestSimpleGet(t *testing.T) {
	rr := probe.TestSimple(ir, nil, testIndexCtl)

	if rr.Result().StatusCode != 200 {
		t.Error("Router for index initialized but request returned non 200 response code:",
			rr.Result().StatusCode)
	}
}
