package rubik

import (
	"io/ioutil"
	"reflect"
	"testing"
)

var probe *TestProbe

type Enn struct {
	testPath string
	Name     string
}

func (enn Enn) Entity() interface{} {
	return enn
}

func (enn Enn) Path() string {
	return enn.testPath
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
		testPath: "/",
		Name:     "ashish",
	}

	rr := probe.Test(entity)
	if rr.Result().StatusCode == 200 {
		defer rr.Result().Body.Close()
		b, err := ioutil.ReadAll(rr.Result().Body)
		if err != nil {
			t.Error(err)
			return
		}

		if string(b) != "Woohoo!" {
			t.Errorf("Resp: %s is not wohoo", string(b))
		} else {
			t.Log("IT IS WHOOHOO")
		}
	}
}
