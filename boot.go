package rubik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/rubikorg/rubik/pkg"
)

func boot() error {
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
						// TODO: parse entity and then pass to the controller -- NOT LIKE THIS !!
						var en interface{}
						if route.Entity == nil {
							en = BlankRequestEntity{}
						} else {
							en = route.Entity
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
							handleErrorResponse(err, writer)
							return
						}

						c, ok := resp.(RenderMixin)

						if ok {
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

func bootStatic() {
	if _, err := os.Stat(pkg.GetStaticFolderPath()); err == nil {
		app.mux.ServeFiles("/static/*filepath", http.Dir("./static"))
		pkg.DebugMsg("Booting => /static")
	}
}

func bootGuard() {}

func bootMiddlewares() {}

func bootController() {}

func handleResponse(response interface{}) {}

func handleErrorResponse(err error, writer http.ResponseWriter) {
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

		b, err := parseHTMLTemplate("/Users/codekid/error.html", "errorTpl", stt)
		if err != nil {
			writer.Write([]byte(err.Error()))
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(b)
		return
	}
	writer.Write([]byte(err.Error()))
}
