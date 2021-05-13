package pkg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"
)

func init() {
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
	workspaceConf := WorkspaceConfig{
		ProjectName: "rubiktest",
		Module:      "rubiktest",
		App: []Project{
			{
				Name:      "appOne",
				Path:      "./cmd/appOne",
				Watchable: false,
				Logging: LoggingConfig{
					Path:      "./Go/ink/logs/$service.log",
					ErrorPath: "./Go/ink/logs/$service.error.log",
					Stream:    "file",
					Format:    "$level: (DD/MM/YYYY) $message",
				},
			},
		},
		X: commands,
	}

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	err := enc.Encode(workspaceConf)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filepath.Join("..", "..", "rubik.toml"), buf.Bytes(), 0755)
	if err != nil {
		panic(err)
	}
}

func TestGetTemplateFolderPath(t *testing.T) {
	path := GetTemplateFolderPath()
	pwd, _ := os.Getwd()
	if path != fmt.Sprintf("%s%s%s", pwd, string(os.PathSeparator), "templates") {
		t.Error("GetTemplateFolderPath: wrong path returned from this method, path:", path)
	}
}

func TestGetRubikConfigPath(t *testing.T) {
	path := GetRubikConfigPath()
	pwd, _ := os.Getwd()
	if path != fmt.Sprintf("%s%s%s", pwd, string(os.PathSeparator), "rubik.toml") {
		t.Error("GetRubikConfigPath: wrong path returned from this method, path:", path)
	}
}

func TestOverrideValues(t *testing.T) {
	source := map[string]interface{}{
		"a": 1,
	}

	env := map[string]interface{}{
		"a": 5,
		"b": 2,
	}

	result := OverrideValues(source, env)
	a := result["a"].(int)
	b := result["b"].(int)

	if a != 5 || b != 2 {
		t.Error("map config override is not working properly. final map: ", result)
	}
}

func TestGetStaticPath(t *testing.T) {
	sp := filepath.Join(".", "static")
	p := GetStaticFolderPath()
	if sp != p {
		t.Error("GetStaticFolderPath() did not return correct static path")
	}
}

func TestGetRubikConfig(t *testing.T) {
	conf, err := GetWorkspaceConfig("../rubik.toml")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if reflect.TypeOf(conf).Elem() != reflect.TypeOf(WorkspaceConfig{}) {
		t.Error("GetRubikConfig() did not return type of project Config")
	}

	if conf.ProjectName == "" {
		t.Error("GetRubikConfig() returns empty project name for a given rubik workspace")
	}
}

func TestGetRubikConfig2(t *testing.T) {
	conf, err := GetWorkspaceConfig("../rubik.toml")
	if err != nil {
		t.Error(err.Error())
	}

	if conf.ProjectName != "core" {
		t.Error("GetRubikConfig() not reading the configs properly")
	}
}

// func TestGetRubikConfig3(t *testing.T) {

// }

func TestMakeAndGetCacheDirPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	correctPath := filepath.Join(home, ".rubik")
	p := MakeAndGetCacheDirPath()
	if p == "" || correctPath != p {
		t.Error("MakeAndGetCacheDirPath() did not return correct path. Path:", p)
	}
}

func TestMakeAndGetCacheDirPath2(t *testing.T) {
	home, _ := os.UserHomeDir()
	correctPath := filepath.Join(home, ".rubik")
	// delete cache dir and test again
	// TODO: move contents of cache and run this test
	// and them move contents back instead of deleting
	os.RemoveAll(correctPath)
	p := MakeAndGetCacheDirPath()
	if p == "" || p != correctPath {
		t.Error("MakeAndGetCacheDirPath() did not return correct path. Path:", p)
	}
}

func TestGetErrorHTMLPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	correctPath := filepath.Join(home, ".rubik", "cache", "error.html")
	p := GetErrorHTMLPath()
	if p == "" || p != correctPath {
		t.Error("GetErrorHTMLPath() did not return correct path. Path:", p)
	}
}
