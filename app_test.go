package sketch

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	var someMap map[string]interface{}
	err := Load(&someMap)
	if err != nil {
		t.Error("sketch is throwing an error even when it didn't find default.toml")
	}
}

func TestCreate(t *testing.T) {
	tRouter := Create("/")
	if reflect.ValueOf(tRouter).Type() != reflect.ValueOf(Router{}).Type() {
		t.Error("Create function did not return instance of Router struct")
	}

	if tRouter.basePath != "/" {
		t.Error("Router has wrong base path in it. Value: " + tRouter.basePath)
	}

	tRouter = Create("/base")

	if tRouter.basePath != "/base" {
		t.Error("Router has wrong base path in it. Value: " + tRouter.basePath)
	}
}
func TestFromStorage(t *testing.T) {
	_, err := FromStorage("noSuchFile")
	if err == nil {
		t.Error("FromStorage() did not throw an error when it did not find the file")
	}
}
