package rubik

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var app = &rubik{
	mux:    *httprouter.New(),
	routes: []Route{},

	beforeHooks: []Responder{},
	routeTree: RouteTree{
		RouterList: make(map[string]string),
		Routes:     []RouteInfo{},
	},
}

func Use(routes ...Route) {
	app.routes = append(app.routes, routes...)
}

func BeforeHook(responder Responder) {
	app.beforeHooks = append(app.beforeHooks, responder)
}

func Run() error {
	c, err := app.loadConfig()
	if err != nil {
		return err
	}
	app.config = c

	err = app.boot()
	if err != nil {
		return err
	}

	// === Plugin code begins here ===
	var plugin string
	flag.StringVar(&plugin, "plugin", "", "output backend information to rubik plugin")
	flag.Parse()
	if plugin != "" {
		fmt.Println("Running plugin:", plugin)
		err := app.streamPluginData()
		if err != nil {
			return err
		}
		return nil
	}

	return http.ListenAndServe(":80", &app.mux)
}
