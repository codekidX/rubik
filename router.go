package rubik

// Router is used to hold all your rubik routes together
type Router struct {
	basePath    string
	routes      []Route
	Middleware  []Controller
	Description string
}

// Add injects a rubik.Route definition to the parent router from which it is called
func (ro *Router) Add(r Route) {
	ro.routes = append(ro.routes, r)
}
