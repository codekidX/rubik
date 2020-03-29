package rubik

import (
	"reflect"
	"testing"
)

func TestGetTestClient(t *testing.T) {
	probe := GetTestClient()
	if reflect.TypeOf(probe).Elem() != reflect.TypeOf(Client{}) {
		t.Error("Probe did not return a value of type rubik.Client")
	}
}
