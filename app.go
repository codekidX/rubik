package cherry

import (
	"errors"
	"net/http"
)

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
	Pattern string
	Handler http.Handler
}

type Middleware func(req Request, next NextFunc)

// App is the instance of App
type Cherry struct {
	State     interface{}
	Config    Config
	mux       *http.ServeMux
	routers   []Router
	routeInfo []RouteInfo
	emitters  map[string]EmitterFunc
}

// CherryPayloadConnector ...
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
	basePath string
	routes   []Route
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

func (a *Cherry) Create(index string) Router {
	return Router{
		basePath: index,
	}
}

func (c *Cherry) Plug(plugin Plugin) {
	c.mux.Handle(plugin.Pattern, plugin.Handler)
}

func (s *Cherry) AddEmitter(event string, efunc EmitterFunc) {
	s.emitters[event] = efunc
}

func (s *Cherry) Emit(event string) error {
	eFunc := s.emitters[event]
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
	return http.ListenAndServe(args[0], c.mux)
}

func (ro *Router) Add(r Route) {
	ro.routes = append(ro.routes, r)
}
