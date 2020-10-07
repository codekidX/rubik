package rubik

import (
	"reflect"
	"testing"
)

var probe *TestProbe

type Enn struct {
	Entity
	Name string
}

func (enn Enn) CoreEntity() interface{} {
	return enn
}

func (enn Enn) ComposedEntity() Entity {
	return enn.Entity
}

func (enn Enn) Path() string {
	return enn.Entity.PointTo
}

func init() {
	probe = NewProbe(initTestRouter())
}

var ir = Route{
	Method:     "GET",
	Path:       "/",
	Entity:     Enn{},
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
	if reflect.TypeOf(probe) != reflect.TypeOf(&TestProbe{}) {
		t.Error("Probe did not return a value of type rubik.TestProbe")
	}
}

func TestIndexRoute(t *testing.T) {
	entity := Enn{
		Name: "ashish",
	}
	entity.PointTo = "/"

	rr := probe.Test(entity)
	if rr.Result().StatusCode == 200 {
		defer rr.Result().Body.Close()
		resp := rr.Body.String()
		if resp != "Woohoo!" {
			t.Errorf("Resp: %s is not wohoo", resp)
		}
	}
}
