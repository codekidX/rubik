package rubik

import (
	"reflect"
	"testing"
)

type testConfig struct {
	Port string
}

type testBlock struct {
}

func (tb testBlock) OnAttach(app *App) error {
	return nil
}

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

func TestGetConfig(t *testing.T) {
	err := Load(&testConfig{})
	if err != nil {
		t.Error(err.Error())
	}

	conf := GetConfig()
	_, ok := conf.(testConfig)
	if !ok {
		t.Error("GetConfig() not returning config of proper type as set while loading")
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

func TestSetNotFoundHandler(t *testing.T) {
	SetNotFoundHandler(notFoundHandler{})
	// shoukd not panic
	intr := recover()
	if intr != nil {
		err := intr.(error)
		t.Error(err.Error())
	}
}

func TestEFunc(t *testing.T) {
	err := E("some error")
	if err.Error() != "some error" {
		t.Error("E() did not save the error message properly")
	}
}

func TestAttach(t *testing.T) {
	Attach("TestBlock", testBlock{})
	if len(app.blocks) == 0 {
		t.Error("Attach() called but number of blocks inside rubik is still 0")
	}

	// check if blocks with same symbol cannot be attached
	Attach("TestBlock", testBlock{})
	if len(app.blocks) > 1 {
		t.Error("Attach() called second time with same symbol and it got attached")
	}
}

// func TestGetBlock(t *testing.T) {
// 	Attach("TestBlock", testBlock{})
// 	_, ok := GetBlock("TestBlock").(*testBlock)
// 	if !ok {
// 		t.Error("GetBlock() not returning proper block struct of type testBlock{}")
// 	}
// }

func TestBeforeRequest(t *testing.T) {
	BeforeRequest(func(rc *HookContext) {})
	if len(beforeHooks) == 0 {
		t.Error("BeforeRequest() not attached to beforeHooks slice")
	}
}

func TestAfterRequest(t *testing.T) {
	AfterRequest(func(rc *HookContext) {})
	if len(afterHooks) == 0 {
		t.Error("AfterRequest() not attached to beforeHooks slice")
	}
}
