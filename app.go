package cherry

import (
	"errors"
	"net/http"
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

type Config struct {
	GenerateDocs bool
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
	Config    Config
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

func (req Request) DefineRawHandler(path string, handler http.HandlerFunc) {

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

type RouteInfo struct {
	Path        string
	Description string
	Entity      interface{}
	IsJSON      bool
	Responses   map[int]string
}

func (c *Cherry) Create(index string) Router {
	return Router{
		basePath: index,
	}
}

func (c *Cherry) Use(router Router) {
	c.routers = append(c.routers, router)
}

func (c *Cherry) Plug(plugin Plugin) {
	c.mux.Handle(plugin.Pattern, plugin.Handler)
}

func (c *Cherry) PlugFunc(plugin Plugin) {
	c.mux.HandleFunc(plugin.Pattern, plugin.HandlerFunc)
}

func (c *Cherry) AddEmitter(event string, efunc EmitterFunc) {
	c.emitters[event] = efunc
}

func (c *Cherry) Emit(event string) error {
	eFunc := c.emitters[event]
	if eFunc == nil {
		return errors.New("Emitter with event: " + event + " is not registered. Call AddEmitter() to add an emitter function to cherry server.")
	}
	eFunc()
	return nil
}

func (c *Cherry) Listen(args ...string) error {
	//c.routers = append(c.routers, c.Index)
	if c.Config.GenerateDocs {
		// TODO: write code to genereate Swagger Docs
	}

	c.boot()
	return http.ListenAndServe(args[0], c.mux)
}

func (c *Cherry) boot() {
	// write the boot sequence of the server
}

// Add injects a cherry.Route definition to the main http server instance
func (ro *Router) Add(r Route) {
	ro.routes = append(ro.routes, r)
}
