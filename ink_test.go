package ink

import (
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

func TestGetCall(t *testing.T) {
	inkcl := New("http://localhost:12333", time.Second*30)
	_, err := inkcl.Get("/somepath/$", "value").Call()
	if err == nil {
		t.Error("There should be an error because 2 $'s are passed and single param")
	}
}
