package cherry

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	i "github.com/okcherry/cherry/pkg"
)

// App is a singleton instance of cherry server
var App = &Cherry{
	mux:     http.NewServeMux(),
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
	Pattern     string
	Handler     http.Handler
	HandlerFunc http.HandlerFunc
}

type Middleware func(req Request, next NextFunc)

// Cherry is the instance of Server which holds all the necessary information of apis
type Cherry struct {
	State     interface{}
	Config    *i.Config
	mux       *http.ServeMux
	routers   []Router
	routeInfo []RouteInfo
	emitters  map[string]EmitterFunc
}

// Request ...
type Request struct {
	RawRequest *http.Request
	cherry     *Cherry
}

func (req Request) GetRouteInfo() []RouteInfo {
	return req.cherry.routeInfo
}

type Router struct {
	basePath   string
	routes     []Route
	Middleware []Middleware
}

// Route is the culmination of
type Route struct {
	Path                 string
	Description          string
	ResponseDeclarations map[int]string
	JSON                 bool
	Entity               interface{}
	Middlewares          []Middleware
	Validate             func(entity interface{}) (bool, string)
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

// Create ...
func (c *Cherry) Create(index string) Router {
	return Router{
		basePath: index,
	}
}

// Use ...
func (c *Cherry) Use(router Router) {
	c.routers = append(c.routers, router)
}

// Plug ...
func (c *Cherry) Plug(plugin Plugin) {
	c.mux.Handle(plugin.Pattern, plugin.Handler)
}

// PlugFunc ...
func (c *Cherry) PlugFunc(plugin Plugin) {
	c.mux.HandleFunc(plugin.Pattern, plugin.HandlerFunc)
}

// AddEmitter ...
func (c *Cherry) AddEmitter(event string, efunc EmitterFunc) {
	c.emitters[event] = efunc
}

// Emit ...
func (c *Cherry) Emit(event string) error {
	eFunc := c.emitters[event]
	if eFunc == nil {
		return errors.New("Emitter with event: " + event + " is not registered. Call AddEmitter() to add an emitter function to cherry server.")
	}
	eFunc()
	return nil
}

// Listen ...
func (c *Cherry) Listen(args ...string) error {
	// if c.Config.GenerateDocs {
	// TODO: write code to genereate /sdk/tools/swagger.json route
	// }

	err := c.boot()

	if err != nil {
		panic(err)
	}

	var port string
	if c.Config != nil {
		// load port from environ
		if os.Getenv("CHERRY_ENV") != "" {
			// get specific env config
			conf := c.Config.App[os.Getenv("CHERRY_ENV")]
			port = conf.Port
		} else {
			i.DebugMsg("CHERRY_ENV is not specified so using cherry.default config")
			port = ":8000"
		}
		i.CherryMsg("Started on port: " + port[1:])
		return http.ListenAndServe(port, c.mux)
	} else if c.Config == nil && len(args) == 0 {
		port = ":6000"
		i.CherryMsg("Started on port: " + port[1:])
		return http.ListenAndServe(port, c.mux)
	}

	i.CherryMsg("Started on port: " + args[0][1:])
	return http.ListenAndServe(args[0], c.mux)
}

func (c *Cherry) RestError(code int, message string) restError {
	return restError{Code: code, Message: message}
}

func (c *Cherry) boot() error {
	c.checkForConfig()
	var errored bool
	// write the boot sequence of the server
	for _, router := range c.routers {
		for index := 0; index < len(router.routes); index++ {
			route := router.routes[index]
			finalPath := router.basePath + route.Path
			i.DebugMsg("Booting => " + finalPath)

			if route.Entity != nil {
				validEntity := checkIsEntity(route.Entity)
				if !validEntity {
					i.ErrorMsg("Your Entity must extend cherry.RequestEntity struct")
					errored = true
					continue
				}
			} else {
				i.WarnMsg(fmt.Sprintf("Please pass in a RequestEntity for route: %s", route.Path))
				continue
			}

			if route.Controller != nil {
				c.mux.HandleFunc(finalPath, func(writer http.ResponseWriter, request *http.Request) {
					resp, err := route.Controller(route.Entity)
					re, ok := err.(restError)

					// error handling
					if err != nil {
						if ok {
							writer.Header().Set("Content-Type", "application/json")
							writer.WriteHeader(re.Code)
							b, _ := json.Marshal(err)
							writer.Write(b)
							return
						}

						// we now make sure that it is not a normal error without a code
						if err.Error() != "" {
							writer.Header().Set("Content-Type", "application/json")
							writer.WriteHeader(500)
							b, _ := json.Marshal(err.Error())
							writer.Write(b)
						}
					}

					a, _ := resp.(string)
					writer.Write([]byte(a))
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

// Load checks for config.toml and loads all the environment variables
func (c *Cherry) checkForConfig() {
	c.Config = i.GetCherryConfig()
	i.DebugMsg("Loaded Config successfully")
}
