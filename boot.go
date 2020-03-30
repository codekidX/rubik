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

// notFoundHandler is rubik's not found route renderer
type notFoundHandler struct{}

func (nfh notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stt := StackTraceTemplate{
		Msg: "Route " + r.URL.Path + " not found",
	}

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

func boot(isREPLMode bool) error {
	if !isREPLMode {
		handle404Response()
		err := bootBlocks()
		if err != nil {
			pkg.ErrorMsg(err.Error())
			return err
		}
	}

	bootStatic()

	//c.checkForConfig()
	var errored bool
	// write the boot sequence of the server
	for _, router := range app.routers {
		for index := 0; index < len(router.routes); index++ {
			route := router.routes[index]
			finalPath := safeRouterPath(router.basePath) + safeRoutePath(route.Path)

			if !isREPLMode {
				pkg.DebugMsg("Booting => " + finalPath)
			}

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
				// DANGER: code should depend upon route.Method
				app.mux.GET(finalPath,
					func(writer http.ResponseWriter, req *http.Request, ps httprouter.Params) {

						reqCtx := RequestContext{
							Request: req,
							Ctx:     make(map[string]interface{}),
						}

						var en interface{}
						if route.Entity != nil {
							var err error
							en, err = inject(req, ps, route.Entity, route.Validation)
							if err != nil {
								// TODO: injection error must be 400 bad request
								handleErrorResponse(err, writer, &reqCtx)
								return
							}
						}

						go dispatchHooks(beforeHooks, &reqCtx)

						// TODO: this is something i need to think about after addding the
						// speculate.go as the middleware response also needs tp be
						// speculated
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

						var resp interface{}
						bresp := route.Controller(en)

						// set values in request context
						reqCtx.Status = bresp.Status

						// check if error response
						if bresp.Status != http.StatusOK {
							rem := RestErrorMixin{
								Code:    bresp.Status,
								Message: bresp.Error.Error(),
							}

							if bresp.OfType == Type.JSON {
								b, _ := json.Marshal(rem)
								writeResponse(writer, bresp.Status, Content.JSON, b)
							} else {
								writeResponse(writer, bresp.Status, Content.Text,
									[]byte(rem.Message))
							}
							reqCtx.Response = rem
						} else {
							resp = bresp.Data

							switch bresp.OfType {
							case Type.HTML:
								s, _ := resp.(string)
								writeResponse(writer, bresp.Status, Content.HTML, []byte(s))
								break
							case Type.Text:
								s, _ := resp.(string)
								writeResponse(writer, bresp.Status, Content.Text, []byte(s))
								break
							case Type.JSON:
								b, _ := json.Marshal(resp)
								writeResponse(writer, bresp.Status, Content.JSON, b)
								break
							case Type.templateHTML, Type.templateText:
								var conType = Content.HTML
								if bresp.OfType == Type.templateText {
									conType = Content.Text
								}

								b, _ := bresp.Data.([]byte)
								writeResponse(writer, bresp.Status, conType, b)
								break
							case Type.Bytes:
								b, _ := resp.([]byte)
								// TODO: write something about this coersion error
								// if !ok {
								// }
								// TODO: check how to set header for a file byte body
								writeResponse(writer, bresp.Status, Content.Text, b)
								break
							default:
								return
							}

							reqCtx.Response = bresp.Data
						}

						go dispatchHooks(afterHooks, &reqCtx)
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

func writeResponse(w http.ResponseWriter, status int, contype string, body []byte) {
	w.Header().Set(Content.Header, contype)
	w.WriteHeader(status)
	w.Write(body)
}

func bootBlocks() error {
	if len(app.blocks) > 0 {
		for k, v := range app.blocks {
			// make sure that some symbols are reserved, this helps when
			// there are blocks which which are attaching to the internal
			// function of rubik
			if isOneOf(k, "session") {
				return errors.New(fmt.Sprintf("%s is a reserved symbol", k))
			}

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

func bootGuard(
	w http.ResponseWriter, g AuthorizationGuard, placebo *App, headers http.Header) error {
	if g.Require() != "" {
		val := app.intermConfig.Get(g.Require())
		if val == nil {
			msg := fmt.Sprintf("No config object [%s]",
				g.Require())
			return errors.New(msg)
		}
	}

	if g.GetRealm() != "" {
		w.Header().Set("WWW-Authenticate", g.GetRealm())
	}

	return g.Authorize(placebo, headers)
}

func bootMiddlewares() {}

func bootController() {}

func handle404Response() {
	if app.mux.NotFound == nil {
		app.mux.NotFound = notFoundHandler{}
	}
}

func handleResponse(response interface{}) {}

func handleErrorResponse(err error, writer http.ResponseWriter, rc *RequestContext) {
	isDevEnv := true
	if !(app.currentEnv == "" || app.currentEnv == "development") {
		isDevEnv = false
	}

	// TODO: fix this mess and in the end afterHooks
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

func dispatchHooks(hooks []RequestHook, rc *RequestContext) {
	if len(hooks) > 0 {
		for _, h := range hooks {
			h(rc)
		}
	}
}
