package rubik

import (
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	var someMap map[string]interface{}
	err := Load(&someMap)
	if err != nil {
		t.Error("Load() is throwing an error even when it didn't find default.toml")
	}

	err = Load(someMap)
	if err == nil {
		t.Error("Load() did not throw an error when you passed a value type")
	}
}

func TestCreate(t *testing.T) {
	tRouter := Create("/")
	if reflect.ValueOf(tRouter).Type() != reflect.ValueOf(Router{}).Type() {
		t.Error("Create() did not return instance of Router struct")
	}

	if tRouter.basePath != "/" {
		t.Error("Router has wrong base path in it. Value: " + tRouter.basePath)
	}

	tRouter = Create("/base")

	if tRouter.basePath != "/base" {
		t.Error("Router has wrong base path in it. Value: " + tRouter.basePath)
	}
}

func TestUse(t *testing.T) {
	r := Create("/")
	Use(r)
	if len(app.routers) <= 0 {
		t.Error("Use() did not append a router to the routes slice of app")
	}
}

func TestFromStorage(t *testing.T) {
	resp := FromStorage("noSuchFile")
	if resp.Error == nil {
		t.Error("FromStorage() did not throw an error when it did not find the file")
	}
}

func TestRestError(t *testing.T) {
	_, err := RestError(400, "something")
	if reflect.ValueOf(err).Type() != reflect.ValueOf(RestErrorMixin{}).Type() {
		t.Error("RestError() did not return a RestErrorMixin")
		return
	}

	re, _ := err.(RestErrorMixin)
	if re.Code != 400 || re.Message != "something" {
		t.Error("RestError() does not contain proper data:", re)
	}
}
