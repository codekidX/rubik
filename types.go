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
	Routes     []*RouteInfo
}

type appConfig struct {
	Cmd  string
	Port int
}

// RouteInfo is a flat structure for processing information about the routes
type RouteInfo struct {
	fullPath  string
	path      string
	doc       string
	groupName string
	method    string
	name      string
}

func (ri *RouteInfo) Name(name string) *RouteInfo {
	ri.name = name
	return ri
}

func (ri *RouteInfo) Doc(doc string) *RouteInfo {
	ri.doc = doc
	return ri
}

func (ri *RouteInfo) Group(name string) *RouteInfo {
	ri.groupName = name
	return ri
}

func (ri *RouteInfo) FullPath(p string) *RouteInfo {
	ri.fullPath = p
	return ri
}

type PluginData struct {
	RouteTree RouteTree
}
