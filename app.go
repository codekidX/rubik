package rubik

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/julienschmidt/httprouter"

	"github.com/BurntSushi/toml"

	"github.com/rubikorg/rubik/pkg"
)

// App is a singleton instance of cherry server
var app = &Rubik{
	mux:     httprouter.New(),
	routers: []Router{},
}

// EmitterFunc defines an anonymous func
type EmitterFunc func()

// NextFunc defines the next middleware function in a series of middleware
type NextFunc func(response interface{})

type restError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (re restError) Error() string {
	return re.Message
}

// Plugin lets you plug middlewares, guards and routes from different modules
type Plugin struct {
	Method  string
	Pattern string
	Handler httprouter.Handle
}

// Middleware intercepts user request and processes it
type Middleware func(req Request, next NextFunc)

// Rubik is the instance of Server which holds all the necessary information of apis
type Rubik struct {
	Config       interface{}
	intermConfig pkg.NotationMap
	mux          *httprouter.Router
	routers      []Router
	routeInfo    []RouteInfo
	emitters     map[string]EmitterFunc
}

// Request ...
type Request struct {
	RawRequest *http.Request
	rubik      *Rubik
}

// GetRouteInfo returns a list of loaded routes in rubik
func (req Request) GetRouteInfo() []RouteInfo {
	return req.rubik.routeInfo
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
		return nil, errors.New("file does not exist")
	}

	b, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, err
	}
	return b, err
}

// GetConfig returns the injected config from the Load method
func GetConfig() interface{} {
	return app.Config
}

// Load method loads the config/RUBIK_ENV.toml file into the interface given
func Load(config interface{}) error {
	configKind := reflect.ValueOf(config).Kind()
	if configKind != reflect.Ptr {
		return errors.New("You need to pass a pointer type to the Load() method found: " + configKind.String())
	}

	var defaultMap map[string]interface{}
	var envMap map[string]interface{}
	var envConfigPath string
	env := os.Getenv("RUBIK_ENV")
	pwd, _ := os.Getwd()
	defaultConfigPath := pwd + string(os.PathSeparator) + "config" +
		string(os.PathSeparator) + "default.toml"
	envConfigNotFound := true

	if env != "" {
		envConfigPath = pwd + string(os.PathSeparator) + "config" +
			string(os.PathSeparator) + env + ".toml"

		if _, err := os.Stat(envConfigPath); os.IsNotExist(err) {
			envConfigNotFound = false
			fmt.Println("setting config not found")
		}
	}
	if envConfigNotFound {
		// if no config files are there inside the config directory we cannot load
		// any config inside the rubik app. so we don't have to error the user
		// giving them the freedom to use rubik without the core feature
		if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
			return nil
		}

		_, err := toml.DecodeFile(defaultConfigPath, config)
		_, err = toml.DecodeFile(defaultConfigPath, &app.intermConfig)

		if err != nil {
			return err
		}
	} else {
		// now we need to override env config values with the default values
		_, err := toml.DecodeFile(defaultConfigPath, &defaultMap)
		_, err = toml.DecodeFile(envConfigPath, &envMap)

		if err != nil {
			return err
		}
		finalMap := pkg.OverrideValues(defaultMap, envMap)
		fmt.Println(finalMap)
		var buf bytes.Buffer
		enc := toml.NewEncoder(bufio.NewWriter(&buf))
		err = enc.Encode(&finalMap)
		if err != nil {
			return err
		}
		err = toml.Unmarshal(buf.Bytes(), config)
		app.intermConfig = finalMap
	}

	app.Config = reflect.ValueOf(config).Elem().Interface()

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
func AddEmitter(event string, efunc EmitterFunc) {
	app.emitters[event] = efunc
}

// Emit ...
func Emit(event string) error {
	eFunc := app.emitters[event]
	if eFunc == nil {
		return errors.New("Emitter with event: " + event +
			" is not registered. Call AddEmitter() to add an emitter function to cherry server.")
	}
	eFunc()
	return nil
}

// Run ...
func Run(args ...string) error {
	err := boot()

	if err != nil {
		panic(err)
	}

	var port string
	if app.Config != nil {
		// load port from environ
		val, err := app.intermConfig.Get("port")
		portVal, ok := val.(string)

		if err != nil || !ok {
			port = ":8000"
		} else {
			port = portVal
		}

		pkg.RubikMsg("Rubik app started on port " + port[1:])
		return http.ListenAndServe(port, app.mux)
	} else if app.Config == nil && len(args) == 0 {
		port = ":8000"
		pkg.RubikMsg("Rubik app started on port " + port[1:])
		return http.ListenAndServe(port, app.mux)
	} else {
		port = ":8000"
	}

	pkg.RubikMsg("Rubik app started on port " + args[0][1:])
	return http.ListenAndServe(args[0], app.mux)
}

// RestError returns a json with the error code and the message
func RestError(code int, message string) restError {
	return restError{Code: code, Message: message}
}

func boot() error {
	//c.checkForConfig()
	var errored bool
	// write the boot sequence of the server
	for _, router := range app.routers {
		for index := 0; index < len(router.routes); index++ {
			route := router.routes[index]

			finalPath := safeRouterPath(router.basePath) + safeRoutePath(route.Path)

			pkg.DebugMsg("Booting => " + finalPath)

			if route.Entity != nil {
				validEntity := checkIsEntity(route.Entity)
				if !validEntity {
					pkg.ErrorMsg("Your Entity must extend cherry.RequestEntity struct")
					errored = true
					continue
				}
			}

			if route.Controller != nil {
				app.mux.GET(finalPath,
					func(writer http.ResponseWriter, req *http.Request, ps httprouter.Params) {
						// TODO: parse entity and then pass to the controller -- NOT LIKE THIS !!
						var en interface{}
						if route.Entity == nil {
							en = BlankRequestEntity{}
						} else {
							en = route.Entity
						}
						resp, err := route.Controller(en)
						re, ok := err.(restError)

						// error handling
						if err != nil {
							if ok {
								writer.Header().Set("Content-Type", "application/json")
								writer.WriteHeader(re.Code)
								b, _ := json.Marshal(err)
								_, _ = writer.Write(b)
								return
							}

							// we now make sure that it is not a normal error without a code
							if err.Error() != "" {
								writer.Header().Set("Content-Type", "application/json")
								writer.WriteHeader(500)
								e := restError{
									Code:    500,
									Message: err.Error(),
								}
								b, _ := json.Marshal(e)
								_, _ = writer.Write(b)
								return
							}
						}

						a, ok := resp.(string)
						if ok {
							_, _ = writer.Write([]byte(a))
							return
						}

						b, ok := resp.([]byte)
						if ok {
							_, _ = writer.Write(b)
						}
						//validReq := route.Validate()
						//if validReq {
						//
						//}
					})
			} else {
				pkg.WarnMsg("ROUTE_NOT_BOOTED - There is no controller assigned for route: " + finalPath)
			}
		}
	}

	if errored {
		return errors.New("encountered following errors while running cherry boot sequence")
	}
	return nil
}

// Load checks for config.toml and loads all the environment variables
func checkForConfig() {
	app.Config = pkg.GetRubikConfig()
	pkg.DebugMsg("Loaded Config successfully")
}
