package rubik

// Router is used to hold all your rubik routes together
type Router struct {
	basePath   string
	routes     []Route
	Middleware []Middleware
}

// Add injects a cherry.Route definition to the main http server instance
func (ro *Router) Add(r Route) {
	ro.routes = append(ro.routes, r)
}

// StorageRoutes create routes inside router that links your storage/fileName
// to the Router base path
func (ro *Router) StorageRoutes(fileNames ...string) {
	for _, file := range fileNames {
		r := Route{
			Method: "GET",
			Path:   safeRoutePath(file),
			Controller: func(entity interface{}) ByteResponse {
				return FromStorage(file)
			},
		}
		ro.routes = append(ro.routes, r)
	}
}
