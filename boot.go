package rubik

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/julienschmidt/httprouter"
)

func (app *rubik) loadConfig() (*appConfig, error) {
	return nil, nil
}

func (app *rubik) boot() error {
	for _, route := range app.routes {
		for _, m := range route.Method {
			// FIXME: we can save some time here if we ignore join here and
			// concat inside this loop itself
			app.routeTree.RouterList[route.Path] = strings.Join(route.Method, "|")
			app.routeTree.Routes = append(app.routeTree.Routes, RouteInfo{
				Path:        route.Path,
				Description: route.Doc,
				Method:      m,
			})
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
			AfterChan: make(chan struct{}),

			writer: w,
			extras: make(map[string]any),
			mu:     &sync.RWMutex{},
		}

		// first run all before hooks
		for _, bh := range app.beforeHooks {
			bh(&rc)
		}
		// we come to the path responders
		for _, responder := range route.Responders {
			responder(&rc)
			if rc.written {
				rc.AfterChan <- struct{}{}
				close(rc.AfterChan)
				break
			}
		}
	}
}

func (app *rubik) streamPluginData() error {
	c, err := net.Dial("unix", rubikSock)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	fmt.Println(app.routeTree)
	err = enc.Encode(PluginData{RouteTree: app.routeTree})
	if err != nil {
		return err
	}

	c.Write(buf.Bytes())
	return nil
}
