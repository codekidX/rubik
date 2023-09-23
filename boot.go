package rubik

import (
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

func (app *rubik) loadConfig() (*appConfig, error) {
	return nil, nil
}

func (app *rubik) boot() error {
	for _, route := range app.routes {
		for _, m := range route.Method {
			switch m {
			case http.MethodGet:
				app.mux.GET(route.Path, executor(route))
			case http.MethodPost:
				app.mux.POST(route.Path, executor(route))
			case http.MethodPut:
				app.mux.PUT(route.Path, executor(route))
			case http.MethodDelete:
				app.mux.DELETE(route.Path, executor(route))
			case http.MethodPatch:
				app.mux.PATCH(route.Path, executor(route))
			case http.MethodOptions:
				app.mux.OPTIONS(route.Path, executor(route))
			default:
				continue
			}
		}
	}
	return nil
}

func executor(route Route) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		rc := Context{
			Request:   r,
			Writer:    w,
			AfterChan: make(chan struct{}),

			extras: make(map[string]any),
			mu:     &sync.RWMutex{},
		}
		for _, responder := range route.Responders {
			responder(&rc)
			if rc.written {
				rc.AfterChan <- struct{}{}
				break
			}
		}
	}
}
