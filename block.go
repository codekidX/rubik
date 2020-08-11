package rubik

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Block is an interface that can be implemented to provide
// extended functionalities to rubik server
// Think of it as a plugin which can be attached to the
// rubik server and can be accessible throughout the
// lifecycle of rubik server.
//
// A Block can also be thought of as a dependency injected
// plugin and can be accessed in your controllers by
// calling rubik.GetBlock('BLOCK_NAME').
// Blocks requires you to implement a method called
// OnAttach. This method is called during rubik server
// bootstrapper is run and requires you to return an error
// if any complexity arrises in for your block to function
type Block interface {
	OnAttach(*App) error
}

// ExtensionBlock is executed plugins when RUBIK_ENV = ext.
// Blocks which requires access to server but does need the
// server to run. To run your extention block use
// `okrubik run --ext`
type ExtensionBlock interface {
	OnPlug(*App) error
	Name() string
}

// App is a sandboxed object used by the external blocks of code
// to access some risk-free part of your rubik server
// For example:
// App do not have full access to your project config but it has
// the ability to decode the config that it needs for
// only this block of code to work
type App struct {
	blockName  string
	app        rubik
	CurrentURL string
	RouteTree
}

// Decode decodes the internal rubik server config into the struct
// that you provide. It returns error if the config is not
// unmarshalable OR there if there is no config
// initialized by the given name parameter
func (sb *App) Decode(name string, target interface{}) error {
	// check for target is pointer or not
	val := sb.app.intermConfig.Get(name)
	msg := fmt.Sprintf("AppDecodeError: block =[ %s ]= requires you to specify "+
		"%s object inside your config/.toml file", sb.blockName, name)
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
