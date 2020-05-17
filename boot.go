package rubik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/printzero/tint"
	"github.com/rubikorg/rubik/pkg"
)

// notFoundHandler implements http.Handler interface
// it shows the error response as stacktrace and
// decieds not to show on non-production env
type notFoundHandler struct{}

// ServeHTTP is the implementation method of http.Handler
func (nfh notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stt := stackTraceTemplate{
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

// boot is the bootstrapper function of rubik server
// it helps to take care of building all the functional
// component and initializing them to make a working
// server
// The sequence of booting is as follows:
//
// 1. bootMessageListener()
// 2. handle404Response()
// 3. bootBlocks()
// 4. bootStatic()
// 5. bootRoutes()
func boot(isREPLMode bool) error {
	go bootMessageListener()

	if !isREPLMode {
		handle404Response()
		err := bootBlocks(app.blocks)
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
		// insert in tree
		app.routeTree.RouterList[strings.ReplaceAll(router.basePath, "/", "")] = router.Description

		for index := 0; index < len(router.routes); index++ {
			route := router.routes[index]
			finalPath := safeRouterPath(router.basePath) + safeRoutePath(route.Path)

			// insert in tree
			rinfo := RouteInfo{
				BelongsTo:   strings.ReplaceAll(router.basePath, "/", ""),
				Entity:      route.Entity,
				Description: route.Description,
				Path:        finalPath,
				Method:      route.Method,
				Responses:   route.ResponseDeclarations,
			}
			app.routeTree.Routes = append(app.routeTree.Routes, rinfo)

			if !isREPLMode {
				pkg.DebugMsg("Booting => " + finalPath)
			}

			if route.Entity != nil {
				if reflect.TypeOf(route.Entity).Kind() != reflect.Ptr {
					return errors.New("Entity field must be a pointer to your Entity")
				}

				validEntity := checkIsEntity(route.Entity)
				if !validEntity {
					pkg.ErrorMsg("Your Entity must extend cherry.Entity struct")
					errored = true
					continue
				}
			}

			handler := func(writer http.ResponseWriter, req *http.Request, ps httprouter.Params) {
				defer req.Body.Close()
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
							Raw:            req,
							Params:         ps,
							ResponseHeader: writer.Header(),
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
			}

			if route.Controller != nil {
				if route.Method == "" {
					app.mux.GET(finalPath, handler)
				} else if !strings.Contains(route.Method, "|") {
					app.mux.Handle(route.Method, finalPath, handler)
				} else {
					methods := strings.Split(route.Method, "|")
					for _, m := range methods {
						app.mux.Handle(m, finalPath, handler)
					}
				}
			} else {
				pkg.WarnMsg("ROUTE_NOT_BOOTED: No controller assigned for route: " + finalPath)
			}
		}
	}

	if !isREPLMode {
		err := bootBlocks(app.afterBlocks)
		if err != nil {
			return err
		}
	}

	if errored {
		return errors.New("BootError: error while running rubik boot sequence")
	}
	return nil
}

// writeResponse is a generic utility function to set the response
// of a request in the normalized state with []byte as parameter
// and set the incoming type with the status
func writeResponse(w http.ResponseWriter, status int, contype string, body []byte) {
	w.Header().Set(Content.Header, contype)
	w.WriteHeader(status)
	w.Write(body)
}

// bootBlocks initializes all the attached blocks and calls
// the onAttach method to boot it's requirements.
// A block is said to be attached only if the return error
// value is nil
func bootBlocks(blockList map[string]Block) error {
	if len(blockList) > 0 {
		for k, v := range blockList {
			// Attaching inherent blocks to the rubik server
			// Inherent blocks are nothing but blocks that
			// are intended to be attached to the core
			// functionality of rubik server
			// Rather than attaching a block as an extended
			// functionality you can access these blocks
			// as a part of core API
			genBlockName := strings.ToLower(k)
			sb := &App{
				app: *app, blockName: k,
				CurrentURL: app.url,
				RouteTree:  app.routeTree,
			}
			if strings.Contains(genBlockName, "messagepasser") {
				msgPasser, ok := v.(Communicator)
				if !ok {
					return errors.New("Value for message passer block must implement the " +
						"MessagePasser interface")
				}
				err := v.OnAttach(sb)
				if err != nil {
					return err
				}
				app.comm[k] = msgPasser
			} else {
				err := v.OnAttach(sb)
				if err != nil {
					return err
				}
			}

			msg := fmt.Sprintf("=[ @(%s) ]= block attached", k)
			msg = tint.Init().Exp(msg, tint.Cyan.Bold())
			fmt.Println(msg)
		}
	}
	return nil
}

// bootStatic boots the ServeFiles handler httprouter
// this functions boots /static route as its index
// and points to the static directory inside this
// project
func bootStatic() {
	if _, err := os.Stat(pkg.GetStaticFolderPath()); err == nil {
		app.mux.ServeFiles("/static/*filepath", http.Dir("./static"))
		pkg.DebugMsg("Booting => /static")
	}
}

// bootGuard is a method that verifies the following this
//
// -> calls guard.Require() to check if the config
// requirement is met by the implementer
// -> calls guard.GetRealm() to check if the client
// needs to be aware of an authorization method or not
// -> calls guard.Authorize() only after evaluating
// both the above function
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

// handle404Response boots the notfounfHandler as mux.NotFound Handler
func handle404Response() {
	if app.mux.NotFound == nil {
		app.mux.NotFound = notFoundHandler{}
	}
}

func handleResponse(response interface{}) {}

// TODO: make this cleaner and better
// this method is used to write error stacktrace response if env is
// dev and not if otherwise
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
		var stt = stackTraceTemplate{
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

// dispatchHooks just calls all the hooks passed as the argument
// this is generally used to call before/after request hooks
// and is intended to be executed as a goroutine
func dispatchHooks(hooks []RequestHook, rc *RequestContext) {
	if len(hooks) > 0 {
		for _, h := range hooks {
			h(rc)
		}
	}
}

// bootMessageListener scaffolds a reciever for inherent rubik channels
// currently used channels for rubik message passing are
//
// 1. Dispatch,Message
// 2. Dispatch.Error
func bootMessageListener() {
	select {
	case msg := <-Dispatch.Message:
		tag := T(msg.Communicator, msg.Topic)
		if app.msgRegistry[tag].ctl == nil {
			msg := fmt.Sprintf("A message was dispatched for %s block, but no receiver found", tag)
			pkg.WarnMsg(msg)
		} else {
			go app.msgRegistry[tag].ctl(msg.Body)
		}
		break
	case err := <-Dispatch.Error:
		pkg.ErrorMsg(err.Error())
		break
	}
}
