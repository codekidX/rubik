package rubik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/printzero/tint"
	"github.com/rubikorg/rubik/pkg"
)

// NotFoundHandler is rubik's not found route renderer
type NotFoundHandler struct{}

func (nfh NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stt := StackTraceTemplate{
		Msg: "Route " + r.URL.Path + " not found",
	}

	var b []byte
	b, err := parseHTMLTemplate(pkg.GetErrorHTMLPath(), "errortmpl", stt)
	if err != nil {
		serr, _ := err.(tracer)
		for _, f := range serr.StackTrace() {
			stt.Stack = append(stt.Stack, fmt.Sprintf("%+s:%d\n", f, f))
		}
		b, _ = parseHTMLTemplate(pkg.GetErrorHTMLPath(), "errortmpl", stt)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(404)
	w.Write(b)
}

func boot() error {
	handle404Response()
	err := bootBlocks()
	if err != nil {
		return err
	}

	bootStatic()

	//c.checkForConfig()
	var errored bool
	// write the boot sequence of the server
	for _, router := range app.routers {
		for index := 0; index < len(router.routes); index++ {
			route := router.routes[index]

			finalPath := safeRouterPath(router.basePath) + safeRoutePath(route.Path)

			pkg.DebugMsg("Booting => " + finalPath)

			if route.Entity != nil {
				if reflect.TypeOf(route.Entity).Kind() != reflect.Ptr {
					return errors.New("Entity field must be a pointer to your RequestEntity")
				}

				validEntity := checkIsEntity(route.Entity)
				if !validEntity {
					pkg.ErrorMsg("Your Entity must extend cherry.RequestEntity struct")
					errored = true
					continue
				}
			}

			if route.Controller != nil {
				app.mux.GET(finalPath,
					func(writer http.ResponseWriter, req *http.Request, ps httprouter.Params) {
						reqCtx := RequestContext{
							Request: req,
							Ctx:     make(map[string]interface{}),
						}

						en, err := inject(req, ps, route.Entity, route.Validation)
						if err != nil {
							handleErrorResponse(err, writer, reqCtx)
							return
						}

						go dispatchHooks(beforeHooks, reqCtx)

						if len(route.Middlewares) > 0 {
							fmt.Println("mw injection")
							for _, m := range route.Middlewares {
								r := Request{
									Raw:    req,
									Params: ps,
								}
								intf := m(r)
								fmt.Println(intf)
							}
						}

						resp, err := route.Controller(en)
						re, ok := err.(RestErrorMixin)

						// error handling
						if err != nil {
							if ok {
								writer.Header().Set("Content-Type", "application/json")
								writer.WriteHeader(re.Code)
								b, _ := json.Marshal(err)
								_, _ = writer.Write(b)
								return
							}

							// we now make sure that it is not a normal error without a code
							handleErrorResponse(err, writer, reqCtx)
							return
						}

						c, ok := resp.(RenderMixin)

						if ok {
							// TODO: add switch statement for type and fix this mess

							writer.Header().Set("Content-Type", c.contentType)
							writer.Write(c.content)
							return
						}

						a, ok := resp.(string)
						if ok {
							_, _ = writer.Write([]byte(a))
							return
						}

						b, ok := resp.([]byte)
						if ok {
							_, _ = writer.Write(b)
						}
						//validReq := route.Validate()
						//if validReq {
						//
						//}
					})
			} else {
				pkg.WarnMsg("ROUTE_NOT_BOOTED: No controller assigned for route: " + finalPath)
			}
		}
	}

	if errored {
		return errors.New("BootError: error while running rubik boot sequence")
	}
	return nil
}

func bootBlocks() error {
	if len(app.blocks) > 0 {
		for k, v := range app.blocks {
			err := v.OnAttach(&App{app: *app, BlockName: k})
			if err != nil {
				return err
			}
			msg := fmt.Sprintf("=[ @(%s) ]= block attached", k)
			msg = tint.Init().Exp(msg, tint.Cyan.Bold())
			fmt.Println(msg)
		}
	}
	return nil
}

func bootStatic() {
	if _, err := os.Stat(pkg.GetStaticFolderPath()); err == nil {
		app.mux.ServeFiles("/static/*filepath", http.Dir("./static"))
		pkg.DebugMsg("Booting => /static")
	}
}

func bootGuard() {}

func bootMiddlewares() {}

func bootController() {}

func handle404Response() {
	if app.mux.NotFound == nil {
		app.mux.NotFound = NotFoundHandler{}
	}
}

func handleResponse(response interface{}) {}

func handleErrorResponse(err error, writer http.ResponseWriter, rc RequestContext) {
	isDevEnv := true
	if !(app.currentEnv == "" || app.currentEnv == "development") {
		isDevEnv = false
	}

	writer.WriteHeader(500)
	if err.Error() != "" && isDevEnv {
		serr, ok := err.(tracer)
		var msg = err.Error()
		var stack []string
		var stt = StackTraceTemplate{
			Msg: msg,
		}
		if ok {
			for _, f := range serr.StackTrace() {
				stack = append(stack, fmt.Sprintf("%+s:%d\n", f, f))
			}
			stt.Stack = stack
		}

		b, err := parseHTMLTemplate(pkg.GetErrorHTMLPath(), "errorTpl", stt)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		rc.Response = b
		rc.Status = 500
		go dispatchHooks(afterHooks, rc)
		writer.Write(b)
		return
	}
	writer.Write([]byte(err.Error()))
}

func dispatchHooks(hooks []RequestHook, rc RequestContext) {
	if len(hooks) > 0 {
		for _, h := range afterHooks {
			h(rc)
		}
	}
}
