package pkg

import (
	"fmt"
	"os"
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
