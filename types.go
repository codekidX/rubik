package rubik

import "github.com/julienschmidt/httprouter"

type rubik struct {
	mux         httprouter.Router
	routes      []Route
	config      *appConfig
	beforeHooks []Responder
}

type Responder func(c *Context)

type Route struct {
	Path       string
	Method     []string
	Responders []Responder
}

// RouteTree represents your routes as a local map for
// getting information about your routes
type RouteTree struct {
	RouterList map[string]string
	Routes     []Route
}

type appConfig struct {
	Cmd  string
	Port int
}

type Router struct {
	prefix string
}
