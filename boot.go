package rubik

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/valyala/fasthttp"
)

func (app *rubik) loadConfig() (*AppConfig, error) {
	env := os.Getenv("RUBIK_ENV")
	appame := os.Getenv("RUBIK_APP")
	if env == "" {
		env = "development"
	}
	configPath := filepath.Join(".", "cmd", appame, "config", "envs", env+".json")
	if f, _ := os.Stat(configPath); f == nil {
		return nil, fmt.Errorf("did not find %s in config folder", configPath)
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var appConfig AppConfig

	err = json.Unmarshal(b, &appConfig)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func load(path, method string, responders []Responder) *RouteInfo {
	ri := RouteInfo{
		path:   path,
		method: method,
	}
	app.routeTree.RouterList[path] = path
	app.routeTree.Routes = append(app.routeTree.Routes, &ri)
	if app.routes == nil {
		app.routes = ResponderMap{
			method: {
				path: responders,
			},
		}
	} else {
		app.routes[method][path] = responders
	}

	return &ri
}

func executor(ftx *fasthttp.RequestCtx, responders []Responder) {
	rc := Context{
		fasthttp:  ftx,
		AfterChan: make(chan struct{}),

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

func getWorkspaceConfig() (*AppConfig, error) {
	var config AppConfig
	wsPath := filepath.Join(".", "rubik.json")
	if f, _ := os.Stat(wsPath); f == nil {
		return nil, errors.New("not a rubik project")
	}
	b, err := os.ReadFile(wsPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func streamPluginData() error {
	wsconf, err := getWorkspaceConfig()
	if err != nil {
		return err
	}
	c, err := net.Dial("unix", rubikSock)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(PluginData{
		WorkspaceConfig: *wsconf,
		RouteTree:       app.routeTree,
	})
	if err != nil {
		return err
	}

	c.Write(buf.Bytes())
	return nil
}
