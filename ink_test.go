package ink

import (
	"fmt"
	"testing"
	"time"
)

func TestParamSubstitution(t *testing.T) {
	inkcl := New("http://localhost:12333", time.Second*30)
	_, err := inkcl.Get("/somepath/$/$", "value").Call()
	if err == nil {
		t.Error("There should be an error because 2 $'s are passed and single param")
	}
}

type SomeResponse struct {
	Tooth string `json:"tooth"`
}

func TestGetCall(t *testing.T) {
	inkcl := New("http://localhost:12333", time.Second*30)
	inkrp := inkcl.Get("/somepath/$", "value")
	inkrp.Query.Set("something", "value")
	inkrp.Query.Set("something2", "value2")
	resp, _ := inkrp.Call()
	fmt.Println(resp.Status)
	fmt.Println(resp.StringBody)
	// if err != nil {
	// 	t.Error(err)
	// }
}
