package cherry

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
)

// App is a singleton instance of cherry server
var App = &Cherry{
	mux: http.NewServeMux(),
	routers: []Router{
		{
			basePath: "/",
		},
	},
}

type EmitterFunc func()

type NextFunc func(response interface{})

type Plugin struct {
	Pattern     string
	Handler     http.Handler
	HandlerFunc http.HandlerFunc
}

type Middleware func(req Request, next NextFunc)

// Cherry is the instance of Server which holds all the necessary information of apis
type Cherry struct {
	State     interface{}
	Config    *Config
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
	Validate             func(entity interface{}) bool
	Controller           func(entity interface{}) interface{}
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
			switch os.Getenv("CHERRY_ENV") {
			case "development":
				port = c.Config.Cherry.Default.Port
				break
			case "production":
				port = c.Config.Cherry.Production.Port
				break
			}
		} else {
			debugMsg("CHERRY_ENV is not specified so using cherry.default config")
			port = c.Config.Cherry.Default.Port
		}
		cherryMsg("Started on port: " + port[1:])
		return http.ListenAndServe(port, c.mux)
	} else if c.Config == nil && len(args) == 0 {
		port = ":6666"
		cherryMsg("Started on port: " + port[1:])
		return http.ListenAndServe(port, c.mux)
	}

	cherryMsg("Started on port: " + args[0][1:])
	return http.ListenAndServe(args[0], c.mux)
}

func (c *Cherry) boot() error {
	c.checkForConfig()
	var errored bool
	// write the boot sequence of the server
	for _, router := range c.routers {
		ourPath := router.basePath
		for _, route := range router.routes {
			finalPath := ourPath + route.Path
			debugMsg("Booting => " + finalPath)

			if route.Entity != nil {
				validEntity := checkIsEntity(route.Entity)
				if !validEntity {
					errorMsg("Your Entity must extend cherry.RequestEntity struct")
					errored = true
					continue
				}
			} else {
				warnMsg(fmt.Sprintf("Please pass in a RequestEntity for route: %s", route.Path))
				continue
			}

			c.mux.HandleFunc(finalPath, func(writer http.ResponseWriter, request *http.Request) {

				//validReq := route.Validate()
				//if validReq {
				//
				//}
			})
		}
	}

	if errored {
		return errors.New("Encountered following errors while running cherry boot sequence")
	}
	return nil
}

// Add injects a cherry.Route definition to the main http server instance
func (ro *Router) Add(r Route) {
	ro.routes = append(ro.routes, r)
}

// Load checks for config.toml and loads all the environment variables
func (c *Cherry) checkForConfig() {
	dir, _ := os.Getwd()
	configPath := dir + "/cherry.toml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		debugMsg("Did not find cherry.toml. Booting server without one.")
		return
	}

	var config Config
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		warnMsg("cherry.toml was found but could not parse it. Error: " + err.Error())
		return
	}

	c.Config = &config
	debugMsg("Loaded Config successfully")
}
