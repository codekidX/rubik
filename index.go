package rubik

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/valyala/fasthttp"
)

var app = &rubik{
	beforeHooks: []Responder{},
	routeTree: RouteTree{
		RouterList: make(map[string]string),
		Routes:     []*RouteInfo{},
	},
}

func Hook(responder Responder) {
	app.beforeHooks = append(app.beforeHooks, responder)
}

func GET(path string, responders ...Responder) *RouteInfo {
	return load(path, http.MethodGet, responders)
}

func POST(path string, responders ...Responder) *RouteInfo {
	return load(path, http.MethodGet, responders)
}

func PUT(path string, responders ...Responder) *RouteInfo {
	return load(path, http.MethodPut, responders)
}

func DELETE(path string, responders ...Responder) *RouteInfo {
	return load(path, http.MethodDelete, responders)
}

func PATCH(path string, responders ...Responder) *RouteInfo {
	return load(path, http.MethodPatch, responders)
}

func OPTIONS(path string, responders ...Responder) *RouteInfo {
	return load(path, http.MethodOptions, responders)
}

func Run() error {
	// c, err := app.loadConfig()
	// if err != nil {
	// 	return err
	// }
	// app.config = c

	// === Plugin code begins here ===
	var plugin string
	flag.StringVar(&plugin, "plugin", "", "output backend information to rubik plugin")
	flag.Parse()
	if plugin != "" {
		fmt.Println("Running plugin:", plugin)
		err := streamPluginData()
		if err != nil {
			return err
		}
		return nil
	}

	return fasthttp.ListenAndServe(":80", func(ctx *fasthttp.RequestCtx) {
		executor(ctx, app.routes[string(ctx.Method())][string(ctx.Path())])
	})
}
