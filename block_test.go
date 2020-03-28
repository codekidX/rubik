package rubik

import (
	"testing"

	"github.com/rubikorg/blocks/ds"
)

func TestAppConfig(t *testing.T) {
	copy := &App{}
	conf := copy.Config("ashish.test")
	if conf != nil {
		t.Error("App.Config did not return nil when you used a dot accessor")
	}
}

func TestAppConfig2(t *testing.T) {
	app.intermConfig = ds.NewNotationMap()
	app.intermConfig.Assign(map[string]interface{}{
		"a": 1,
	})

	copy := &App{
		app: *app,
	}
	conf := copy.Config("a")
	if conf != 1 {
		t.Error("App.Config did not return value 1 accessing a config")
	}
}
