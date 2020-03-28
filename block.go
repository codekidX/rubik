package rubik

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Block is a guideline for extended functionality in rubik
type Block interface {
	OnAttach(*App) error
}

// App is a restricted environment for blocks development
type App struct {
	BlockName string
	app       rubik
}

// Decode ..injects your named configs
// sb curresponds to sandboxed app
func (sb *App) Decode(name string, target interface{}) error {
	// check for target is pointer or not
	val := sb.app.intermConfig.Get(name)
	msg := fmt.Sprintf("AppDecodeError: block =[ %s ]= requires you to specify "+
		"%s object inside your config/.toml file", sb.BlockName, name)
	if val == nil {
		return errors.New(msg)
	}

	b, err := json.Marshal(val)
	err = json.Unmarshal(b, target)
	if err != nil {
		return err
	}

	return nil
}

// Config get config by name
func (sb *App) Config(name string) interface{} {
	if strings.Contains(name, ".") {
		return nil
	}

	return sb.app.intermConfig.Get(name)
}
