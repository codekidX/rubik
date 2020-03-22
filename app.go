package rubik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/pkg/errors"

	"github.com/julienschmidt/httprouter"

	"github.com/BurntSushi/toml"

	"github.com/rubikorg/blocks/ds"
	"github.com/rubikorg/rubik/pkg"
)

// App is a singleton instance of rubik server
var app = &rubik{
	mux:     httprouter.New(),
	routers: []Router{},
	logger: &pkg.Logger{
		CanLog: true,
	},
	blocks: make(map[string]Block),
}

var blocks = make(map[string]interface{})
// Session is a manager for managing rubik server sessions
var Session SessionManager

const (
	// Version of rubik
	Version = "v0.1"
)

// EmitterFunc defines an anonymous func
type EmitterFunc func()

type tracer interface {
	StackTrace() errors.StackTrace
}

// RestErrorMixin type is used by rubik to show error in a same format
type RestErrorMixin struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (re RestErrorMixin) Error() string {
	return re.Message
}

// Plugin lets you plug middlewares, guards and routes from different modules
type Plugin struct {
	Method  string
	Pattern string
	Handler httprouter.Handle
}

// Middleware intercepts user request and processes it
type Middleware func(req Request) interface{}

// Rubik is the instance of Server which holds all the necessary information of apis
type rubik struct {
	config       interface{}
	intermConfig ds.NotationMap
	rootConfig   *pkg.Config
	logger       *pkg.Logger
	emitters     map[string]EmitterFunc
	currentEnv   string
	mux          *httprouter.Router
	blocks       map[string]Block
	routers      []Router
	routeInfo    []RouteInfo
}

// Request ...
type Request struct {
	Raw            *http.Request
	Params         httprouter.Params
	ResponseHeader Values
	Session        SessionManager
	app            *rubik
	entity         interface{}
}

// GetRouteInfo returns a list of loaded routes in rubik
func (req Request) GetRouteInfo() []RouteInfo {
	return req.app.routeInfo
}

// Config returns the configuration of your server  for a specific accessor
func (req Request) Config(accessor string) interface{} {
	val := req.app.intermConfig.Get(accessor)
	if val == nil {
		msg := fmt.Sprintf("MiddlewareAccessorError: cannot access %s from project config",
			accessor)
		pkg.ErrorMsg(msg)
		return nil
	}
	return val
}

// Route is the culmination of
type Route struct {
	Path                 string
	Method               string
	Description          string
	ResponseDeclarations map[int]string
	JSON                 bool
	Entity               interface{}
	Middlewares          []Middleware
	Validate             func(entity interface{}) error
	Controller           func(entity interface{}) (interface{}, error)
}

// RouteInfo ...
type RouteInfo struct {
	Path        string
	Description string
	Entity      interface{}
	IsJSON      bool
	Responses   map[int]string
}

// FromStorage returns the file bytes of a given fileName as response
func FromStorage(fileName string) ([]byte, error) {
	pwd, _ := os.Getwd()
	var filePath = pwd + string(os.PathSeparator) + "storage" +
		string(os.PathSeparator) + fileName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("FileNotFoundError: " + fileName + " does not exist.")
	}

	b, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, err
	}
	return b, err
}

// GetConfig returns the injected config from the Load method
func GetConfig() interface{} {
	return app.config
}

// Attach a block to rubik tree
func Attach(symbol string, b Block) {
	if app.blocks[symbol] != nil {
		msg := fmt.Sprintf("Block %s will not be attached on boot as symbol: %s exists", symbol, symbol)
		pkg.ErrorMsg(msg)
		return
	}
	app.blocks[symbol] = b
}

// GetBlock returns the block that is attached to rubik represented by the
// symbol supplied as the parameter
func GetBlock(symbol string) Block {
	return app.blocks[symbol]
}
// Load method loads the config/RUBIK_ENV.toml file into the interface given
func Load(config interface{}) error {
	configKind := reflect.ValueOf(config).Kind()

	if configKind != reflect.Ptr {
		msg := fmt.Sprintf("NonPointerValueError: Load() method requires pointer variable: %s",
			configKind.String())
		return errors.New(msg)
	}

	var defaultMap map[string]interface{}
	var envMap map[string]interface{}
	var envConfigPath string

	env := os.Getenv("RUBIK_ENV")
	// set the current env to app.currentEnv
	app.currentEnv = env

	pwd, _ := os.Getwd()
	defaultConfigPath := pwd + string(os.PathSeparator) + "config" +
		string(os.PathSeparator) + "default.toml"
	envConfigFound := false

	if env != "" {
		envConfigPath = pwd + string(os.PathSeparator) + "config" +
			string(os.PathSeparator) + env + ".toml"

		if _, err := os.Stat(envConfigPath); os.IsNotExist(err) {
			// do this with logger
			msg := fmt.Sprintf("ConfigNotFound: config file %s.toml does not exist", env)
			pkg.DebugMsg(msg)
		} else {
			envConfigFound = true
		}
	}

	app.intermConfig = ds.NewNotationMap()

	if !envConfigFound {
		// if no config files are there inside the config directory we cannot load
		// any config inside the rubik app. so we don't have to error the user
		// giving them the freedom to use rubik without the core feature
		if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
			return nil
		}

		_, err := toml.DecodeFile(defaultConfigPath, config)
		// you can use envMap here since there is no env found and that assignment
		// is not going anywhere until this scope ends we make use of the resources
		_, err = toml.DecodeFile(defaultConfigPath, &envMap)
		if err != nil {
			return errors.WithStack(err)
		}

		app.intermConfig.Assign(envMap)
	} else {
		// now we need to override env config values with the default values
		_, err := toml.DecodeFile(defaultConfigPath, &defaultMap)
		_, err = toml.DecodeFile(envConfigPath, &envMap)
		if err != nil {
			return errors.WithStack(err)
		}

		finalMap := pkg.OverrideValues(defaultMap, envMap)
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		err = enc.Encode(&finalMap)
		if err != nil {
			return errors.WithStack(err)
		}

		err = toml.Unmarshal(buf.Bytes(), config)
		app.intermConfig.Assign(finalMap)
	}

	// irrespective of env found or not flatten the intermconfig
	if app.intermConfig.Length() > 0 {
		app.intermConfig.Flatten()
	}

	app.config = reflect.ValueOf(config).Elem().Interface()

	// before loading anything to interm config mark notation map as not editable
	app.intermConfig.IsEditable(false)

	return nil
}

// Create ...
func Create(index string) Router {
	return Router{
		basePath: index,
	}
}

// Use ...
func Use(router Router) {
	app.routers = append(app.routers, router)
}

//Plug ...
func Plug(plugin Plugin) {
	app.mux.Handle(plugin.Method, plugin.Pattern, plugin.Handler)
}

// PlugAfter ...
func PlugAfter(plugin Plugin) {
	app.mux.Handle(plugin.Method, plugin.Pattern, plugin.Handler)
}

// AddEmitter ...
// func AddEmitter(event string, efunc EmitterFunc) {
// 	app.emitters[event] = efunc
// }

// Emit ...
// func Emit(event string) error {
// 	eFunc := app.emitters[event]
// 	if eFunc == nil {
// 		return errors.New("Emitter with event: " + event +
// 			" is not registered. Call AddEmitter() to add an emitter function to cherry server.")
// 	}
// 	eFunc()
// 	return nil
// }

// SetNotFoundHandler sets custom handler for 404
func SetNotFoundHandler(h http.Handler) {
	app.mux.NotFound = h
}

// Run rubik server instance
func Run(args ...string) error {
	err := boot()

	if err != nil {
		panic(err)
	}

	var port string
	if app.config != nil {
		// load port from environ
		val := app.intermConfig.Get("port")
		portVal, ok := val.(string)
		if val == nil || !ok {
			port = ":8000"
		} else {
			port = portVal
		}

	} else if app.config == nil && len(args) == 0 {
		port = ":8000"
	}

	if port == "" && len(args) > 0 {
		port = args[0]
	}

	pkg.RubikMsg("Rubik server started on port " + port[1:])

	return http.ListenAndServe(port, app.mux)
}

// RestError returns a json with the error code and the message
func RestError(code int, message string) RestErrorMixin {
	return RestErrorMixin{Code: code, Message: message}
}
