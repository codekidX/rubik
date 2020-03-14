package pkg

import (
	"testing"
)

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
