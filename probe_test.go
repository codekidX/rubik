package rubik

import (
	"reflect"
	"testing"
)

var probe *Client

func init() {
	initTestRouter()
	probe = Probe()
}

func initTestRouter() {
	indexRouter := Create("/")
	i := Route{
		Path:       "/",
		Controller: testIndexCtl,
	}
	indexRouter.Add(i)
	Use(indexRouter)
}

func testIndexCtl(req *Request) {
	req.Respond("Woohoo!")
}

func TestGetTestClient(t *testing.T) {
	if reflect.TypeOf(probe).Elem() != reflect.TypeOf(Client{}) {
		t.Error("Probe did not return a value of type rubik.Client")
	}
}

func TestGetCallWithTestClient(t *testing.T) {
	en := BlankRequestEntity{}
	en.PointTo = "/"
	resp, err := probe.Get(en)
	if err != nil {
		t.Error(err.Error())
	}

	if resp.Status != 200 {
		t.Error("Router for index initialized but request returned non 200 response code:",
			resp.Status)
	}
}
