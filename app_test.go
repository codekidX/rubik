package rubik

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/rubikorg/rubik/pkg"
)

type testConfig struct {
	Port string
}

type testBlock struct {
}

func (tb testBlock) OnAttach(app *App) error {
	return nil
}

func getDummyWsConfig() pkg.WorkspaceConfig {
	commands := make(map[string]map[string]string)
	commands["test"] = map[string]string{
		"cwd": "go test -cover ./...",
		"pwd": "./Go/ink",
	}
	commands["storage"] = map[string]string{
		"cwd": "go test -cover storage_test.go",
		"pwd": "./Go/ink",
	}

	// we need to write a rubik.toml file for testing our
	// workspace config
	return pkg.WorkspaceConfig{
		ProjectName: "rubiktest",
		Module:      "rubiktest",
		App: []pkg.Project{
			{
				Name:      "appOne",
				Path:      "./cmd/appOne",
				Watchable: false,
				Logging: pkg.LoggingConfig{
					Path:      "./Go/ink/logs/$service.log",
					ErrorPath: "./Go/ink/logs/$service.error.log",
					Stream:    "file",
					Format:    "$level: (DD/MM/YYYY) $message",
				},
			},
		},
		X: commands,
	}
}

func init() {
	workspaceConf := getDummyWsConfig()

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	err := enc.Encode(workspaceConf)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filepath.Join("..", "..", "rubik.toml"), buf.Bytes(), 0755)
	err = ioutil.WriteFile(filepath.Join(".", "config", "test.toml"), buf.Bytes(), 0755)
	err = ioutil.WriteFile(filepath.Join(".", "config", "default.toml"), buf.Bytes(), 0755)
	if err != nil {
		panic(err)
	}
}

func TestLoad(t *testing.T) {
	var someMap map[string]interface{}
	err := Load(&someMap)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = Load(someMap)
	if err == nil {
		t.Error("Load() did not throw an error when you passed a value type")
	}
}

// func TestGetConfig(t *testing.T) {
// 	err := Load(&testConfig{})
// 	if err != nil {
// 		t.Error(err.Error())
// 		return
// 	}

// 	conf := GetConfig()
// 	if conf == nil {
// 		t.Error("GetConfig() returned nil")
// 		return
// 	}
// 	t.Logf("%+v\n", conf)
// 	_, ok := conf.(testConfig)
// 	if !ok {
// 		t.Error("GetConfig() not returning config of proper type as set while loading")
// 	}
// }

func TestCreate(t *testing.T) {
	tRouter := Create("/")
	if reflect.ValueOf(tRouter).Type() != reflect.ValueOf(Router{}).Type() {
		t.Error("Create() did not return instance of Router struct")
		return
	}

	if tRouter.basePath != "/" {
		t.Error("Router has wrong base path in it. Value: " + tRouter.basePath)
		return
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
	// should not panic
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
		return
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
