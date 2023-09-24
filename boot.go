package rubik

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
)

func (app *rubik) loadConfig() (*appConfig, error) {
	return nil, nil
}

func load(path, method string, responders []Responder) *RouteInfo {
	ri := RouteInfo{
		path:   path,
		method: method,
	}
	app.routeTree.RouterList[path] = path
	app.routeTree.Routes = append(app.routeTree.Routes, &ri)

	switch method {
	case http.MethodGet:
		app.mux.GET(path, executor(path, responders))
	case http.MethodPost:
		app.mux.POST(path, executor(path, responders))
	case http.MethodPut:
		app.mux.PUT(path, executor(path, responders))
	case http.MethodDelete:
		app.mux.DELETE(path, executor(path, responders))
	case http.MethodPatch:
		app.mux.PATCH(path, executor(path, responders))
	case http.MethodOptions:
		app.mux.OPTIONS(path, executor(path, responders))
	default:
		return &ri
	}
	return &ri
}

func executor(path string, responders []Responder) httprouter.Handle {
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
		for _, responder := range responders {
			responder(&rc)
			if rc.written {
				rc.AfterChan <- struct{}{}
				close(rc.AfterChan)
				break
			}
		}
	}
}

func streamPluginData() error {
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
