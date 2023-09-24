package rubik

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

var app = &rubik{
	mux:    *httprouter.New(),
	routes: []Route{},

	beforeHooks: []Responder{},
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

	return http.ListenAndServe(":80", &app.mux)
}
