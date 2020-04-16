package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
	conf := GetRubikConfig()
	if reflect.TypeOf(conf).Elem() != reflect.TypeOf(Config{}) {
		t.Error("GetRubikConfig() did not return type of project Config")
	}

	if conf.ProjectName != "" {
		t.Error("GetRubikConfig() returns some project name eventhough there is no rubik.toml here")
	}
}

func TestGetRubikConfig2(t *testing.T) {
	p, _ := filepath.Abs("..")
	os.Chdir(p)
	conf := GetRubikConfig()

	if conf.ProjectName != "core" {
		t.Error("GetRubikConfig() not reading the configs properly")
	}
}

func TestGetRubikConfig3(t *testing.T) {

}

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
	correctPath := filepath.Join(home, ".rubik", "error.html")
	p := GetErrorHTMLPath()
	if p == "" || p != correctPath {
		t.Error("GetErrorHTMLPath() did not return correct path. Path:", p)
	}
}
