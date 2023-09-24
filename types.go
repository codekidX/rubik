package rubik

import "github.com/julienschmidt/httprouter"

type rubik struct {
	mux         httprouter.Router
	routes      []Route
	config      *appConfig
	beforeHooks []Responder

	routeTree RouteTree
}

type Responder func(c *Context)

type Route struct {
	Path       string
	Method     []string
	Responders []Responder
	Doc        string
}

// RouteTree represents your routes as a local map for
// getting information about your routes
type RouteTree struct {
	RouterList map[string]string
	Routes     []RouteInfo
}

type appConfig struct {
	Cmd  string
	Port int
}

// RouteInfo is a flat structure for processing information about the routes
type RouteInfo struct {
	FullPath    string
	Path        string
	Description string
	BelongsTo   string
	Entity      interface{}
	IsJSON      bool
	Method      string
	Responses   map[int]string
}

type PluginData struct {
	RouteTree RouteTree
}
