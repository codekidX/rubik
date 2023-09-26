package plugin_test

import (
	"testing"

	"github.com/rubikorg/rubik/plugin"
)

func TestGetPluginData(t *testing.T) {
	_, err := plugin.GetPluginData()
	if err != nil {
		t.Error(err)
	}
}
