package sketch

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/julienschmidt/httprouter"

	"github.com/BurntSushi/toml"

	i "github.com/oksketch/sketch/pkg"
)

// App is a singleton instance of cherry server
var app = &Sketch{
	mux:     httprouter.New(),
	routers: []Router{},
}

type EmitterFunc func()

type NextFunc func(response interface{})

type restError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (re restError) Error() string {
	return re.Message
}

type Plugin struct {
	Method  string
	Pattern string
	Handler httprouter.Handle
}

type Middleware func(req Request, next NextFunc)

// Cherry is the instance of Server which holds all the necessary information of apis
type Sketch struct {
	Config       interface{}
	intermConfig i.NotationMap
	mux          *httprouter.Router
	routers      []Router
	routeInfo    []RouteInfo
	emitters     map[string]EmitterFunc
}

// Request ...
type Request struct {
	RawRequest *http.Request
	sketch     *Sketch
}

func (req Request) GetRouteInfo() []RouteInfo {
	return req.sketch.routeInfo
}

type Router struct {
	basePath   string
	routes     []Route
	Middleware []Middleware
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

func FromStorage(fileName string) ([]byte, error) {
	pwd, _ := os.Getwd()
	var filePath = pwd + string(os.PathSeparator) + "storage" + string(os.PathSeparator) + fileName

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, errors.New("file does not exist")
	}

	b, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, err
	}
	return b, err
}

func GetConfig() interface{} {
	return app.Config
}

func Load(config interface{}) error {
	env := os.Getenv("SKETCH_ENV")
	pwd, _ := os.Getwd()
	var configPath string
	if env != "" {
		configPath = pwd + string(os.PathSeparator) + "config" + string(os.PathSeparator) + env + ".toml"
	} else {
		configPath = pwd + string(os.PathSeparator) + "config" + string(os.PathSeparator) + "default.toml"
	}

	_, err := toml.DecodeFile(configPath, config)

	configKind := reflect.ValueOf(config).Kind()
	if configKind != reflect.Ptr {
		return errors.New("You need to pass a pointer type to the Load() method found: " + configKind.String())
	}
	app.Config = reflect.ValueOf(config).Elem().Interface()

	if err != nil {
		return err
	}
	_, err = toml.DecodeFile(configPath, &app.intermConfig)

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

// PlugFunc ...
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
		return errors.New("Emitter with event: " + event + " is not registered. Call AddEmitter() to add an emitter function to cherry server.")
	}
	eFunc()
	return nil
}

// Listen ...
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

		i.SketchMsg("Sketch app started on port " + port[1:])
		return http.ListenAndServe(port, app.mux)
	} else if app.Config == nil && len(args) == 0 {
		port = ":8000"
		i.SketchMsg("Sketch app started on port " + port[1:])
		return http.ListenAndServe(port, app.mux)
	} else {
		port = ":8000"
	}

	i.SketchMsg("Sketch app started on port " + args[0][1:])
	return http.ListenAndServe(args[0], app.mux)
}

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

			i.DebugMsg("Booting => " + finalPath)

			if route.Entity != nil {
				validEntity := checkIsEntity(route.Entity)
				if !validEntity {
					i.ErrorMsg("Your Entity must extend cherry.RequestEntity struct")
					errored = true
					continue
				}
			}

			if route.Controller != nil {
				app.mux.GET(finalPath, func(writer http.ResponseWriter, req *http.Request, ps httprouter.Params) {
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
				i.WarnMsg("ROUTE_NOT_BOOTED - There is no controller assigned for route: " + finalPath)
			}
		}
	}

	if errored {
		return errors.New("encountered following errors while running cherry boot sequence")
	}
	return nil
}

// Add injects a cherry.Route definition to the main http server instance
func (ro *Router) Add(r Route) {
	ro.routes = append(ro.routes, r)
}

// StorageRoutes create routes inside router that links your storage/fileName to the Router base path
func (ro *Router) StorageRoutes(fileNames ...string) {
	for _, file := range fileNames {
		r := Route{
			Method: "GET",
			Path:   safeRoutePath(file),
			Controller: func(entity interface{}) (interface{}, error) {
				return FromStorage(file)
			},
		}
		ro.routes = append(ro.routes, r)
	}
}

// Load checks for config.toml and loads all the environment variables
func checkForConfig() {
	app.Config = i.GetCherryConfig()
	i.DebugMsg("Loaded Config successfully")
}
