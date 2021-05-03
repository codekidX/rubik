// Package rubik is used for accessing Rubik Framework: a minimal and efficient web framework
// for Go and it's APIs.
//
// Running an empty server:
// 		package main
//
// 		import r "github.com/rubikorg/rubik"
//
// 		func main() {
//			// this runs Rubik server on port: 8000
// 			panic(r.Run())
// 		}
//
// Adding a route:
//
// 		package main
//
// 		import r "github.com/rubikorg/rubik"
//
// 		func main() {
//			// this runs Rubik server on port: 8000
// 			index := rubik.Route{
// 				Path: "/",
// 				Controller: func (req *r.Request) { req.Respond("This is a text response") },
// 			}
// 			rubik.UseRoute(index)
// 			panic(r.Run())
// 		}
package rubik

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/julienschmidt/httprouter"

	"github.com/BurntSushi/toml"

	"github.com/rubikorg/blocks/ds"
	"github.com/rubikorg/rubik/pkg"
)

// App is a singleton instance of rubik server
var app = &rubik{
	mux:     httprouter.New(),
	routers: []Router{},
	logger: &pkg.Logger{
		CanLog: true,
	},
	blocks:      make(map[string]Block),
	afterBlocks: make(map[string]Block),
	routeTree: RouteTree{
		RouterList: make(map[string]string),
		Routes:     []RouteInfo{},
	},
	extensions: []Plugin{},
}

var blocks = make(map[string]interface{})
var beforeHooks []RequestHook
var afterHooks []RequestHook

// Dispatch is topic dispatcher of rubik server
var Dispatch = MessagePasser{
	Message: make(chan Message),
	Error:   make(chan error),
}

// Log is a collection of channels of strings which are used to
// stream logs into a folder called logs, where "E" channel
// writes to $app.rubik.error.log and "I" channel writes to
// $app.rubik.info.log file inside the logs/ folder
var Log = struct {
	E chan string
	I chan string
}{
	E: make(chan string),
	I: make(chan string),
}

const (
	// Version of rubik
	Version = "0.2.5"
)

type tracer interface {
	StackTrace() errors.StackTrace
}

// RestErrorMixin type is used by rubik when rubik.Throw is called for
// writing error types as common JSON structure across Rubik server
type RestErrorMixin struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface of Go
func (re RestErrorMixin) Error() string {
	return re.Message
}

// Controller ...
type Controller func(*Request)

// Request ...
type Request struct {
	app     *rubik
	Entity  interface{}
	Session SessionManager
	Writer  RResponseWriter
	Params  httprouter.Params
	Raw     *http.Request
	Ctx     context.Context
	Claims  Claims
}

// Claims populates the JWT.MapClaims inteface
type Claims interface{}

// HookContext ...
type HookContext struct {
	Request  *http.Request
	Ctx      map[string]interface{}
	Response []byte
	Status   int
}

// RequestHook ...
type RequestHook func(*HookContext)

// Rubik is the instance of Server which holds all the necessary information of apis
type rubik struct {
	config       interface{}
	intermConfig ds.NotationMap
	wsConfig     *pkg.WorkspaceConfig
	logger       *pkg.Logger
	currentEnv   string
	url          string
	mux          *httprouter.Router
	blocks       map[string]Block
	afterBlocks  map[string]Block
	routers      []Router
	routeTree    RouteTree
	dep          interface{}
	extensions   []Plugin
}

// GetRouteTree returns a list of loaded routes in rubik
func (req Request) GetRouteTree() RouteTree {
	return req.app.routeTree
}

// Config returns the configuration of your server  for a specific accessor
func (req Request) Config(accessor string) interface{} {
	val := req.app.intermConfig.Get(accessor)
	if val == nil {
		msg := fmt.Sprintf("MiddlewareAccessorError: cannot access %s from "+
			"project config", accessor)
		pkg.ErrorMsg(msg)

		return nil
	}

	return val
}

// Route defines how a specific route route inside the Rubik server must behave.
// Route collects all the information required for processing of a HTTP request
// and performs a handler construction depending upon these values.
//
// There is a specific order in which handlers of Routes are constructed:
//
// [ Entity check --- Guard() --- Validation() --- []Middlewares()
// --- Controller() ]
type Route struct {
	Path                 string
	Method               string
	Description          string
	ResponseDeclarations map[int]string
	JSON                 bool
	Export               bool
	Entity               interface{}
	Guards               []Controller
	Middlewares          []Controller
	Validation           Validation
	Controller           Controller
}

// RouteTree represents your routes as a local map for
// getting information about your routes
type RouteTree struct {
	RouterList map[string]string
	Routes     []RouteInfo
}

// RouteInfo is a flat structure for processing information about the routes
type RouteInfo struct {
	FullPath    string
	Path        string
	Description string
	BelongsTo   string
	Entity      interface{}
	IsJSON      bool
	Method      string
	Responses   map[int]string
}

// GetConfig returns the injected config from the Load method
func GetConfig() interface{} {
	return app.config
}

// Attach a block to rubik tree
func Attach(symbol string, b Block) {
	name := strings.ToLower(symbol)
	if app.blocks[name] != nil {
		msg := fmt.Sprintf("Block %s will not be attached on boot as symbol: %s exists",
			symbol, name)
		pkg.ErrorMsg(msg)

		return
	}

	app.blocks[name] = b
}

// AttachAfter attaches blocks after boot sequence of routes are complete
func AttachAfter(symbol string, b Block) {
	name := strings.ToLower(symbol)
	if app.afterBlocks[name] != nil {
		msg := fmt.Sprintf("Block %s will not be attached on boot as symbol: %s exists",
			symbol, name)
		pkg.ErrorMsg(msg)

		return
	}

	app.afterBlocks[name] = b
}

// GetBlock returns the block that is attached to rubik represented by the
// symbol supplied as the parameter
func GetBlock(symbol string) Block {
	return app.blocks[strings.ToLower(symbol)]
}

// Plug adds an extension of Rubik to your workflow
func Plug(ext Plugin) {
	app.extensions = append(app.extensions, ext)
}

// BeforeRequest is used to execute the request hook h. When a request is sent on a certain route
// the hook specified as h is executed in a separate goroutine without hindering the current
// main goroutine of request.
func BeforeRequest(h RequestHook) {
	beforeHooks = append(beforeHooks, h)
}

// AfterRequest is used to execute the request hook h after completion of the request. A
// request is said to be complete only after the response is written through http.ResponseWriter
// interface of http.Server.
func AfterRequest(h RequestHook) {
	afterHooks = append(afterHooks, h)
}

// Load method loads the config/RUBIK_ENV.toml file into the interface given
func Load(config interface{}) error {
	configKind := reflect.ValueOf(config).Kind()
	if configKind != reflect.Ptr {
		fmtmsg := "NonPointerValueError: Load() method requires pointer variable: %s"
		msg := fmt.Sprintf(fmtmsg, configKind.String())

		return errors.New(msg)
	}

	wsConfig, err := pkg.GetRubikConfig()
	if err != nil {
		return err
	}
	app.wsConfig = wsConfig

	var defaultMap map[string]interface{}
	var envMap map[string]interface{}
	var envConfigPath string

	// set the current env to app.currentEnv
	env := os.Getenv("RUBIK_ENV")
	app.currentEnv = env

	defaultConfigPath := filepath.Join(".", "config", "default.toml")
	envConfigFound := false

	if env != "" && env != "plugin" {
		envConfigPath = filepath.Join(".", "config", env+".toml")

		if _, err := os.Stat(envConfigPath); os.IsNotExist(err) {
			// do this with logger
			msg := fmt.Sprintf("ConfigNotFound: config file %s.toml does not exist",
				env)
			pkg.DebugMsg(msg)
		} else {
			envConfigFound = true
		}
	}

	app.intermConfig = ds.NewNotationMap()

	if !envConfigFound {
		// if no config files are there inside the config directory we cannot load
		// any config inside the rubik app. so we don't have to error the user
		// giving them the freedom to use rubik without the core feature
		if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
			return nil
		}

		_, err := toml.DecodeFile(defaultConfigPath, config)
		// you can use envMap here since there is no env found and that assignment
		// is not going anywhere until this scope ends we make use of the resources
		_, err = toml.DecodeFile(defaultConfigPath, &envMap)
		if err != nil {
			return errors.WithStack(err)
		}

		app.intermConfig.Assign(envMap)
	} else {
		// now we need to override env config values with the default values
		_, err := toml.DecodeFile(defaultConfigPath, &defaultMap)
		_, err = toml.DecodeFile(envConfigPath, &envMap)
		if err != nil {
			return errors.WithStack(err)
		}

		finalMap := pkg.OverrideValues(defaultMap, envMap)
		var buf bytes.Buffer
		enc := toml.NewEncoder(&buf)
		err = enc.Encode(&finalMap)
		if err != nil {
			return errors.WithStack(err)
		}

		err = toml.Unmarshal(buf.Bytes(), config)
		app.intermConfig.Assign(finalMap)
	}

	// irrespective of env found or not flatten the intermconfig
	if app.intermConfig.Length() > 0 {
		app.intermConfig.Flatten()
	}

	app.config = reflect.ValueOf(config).Elem().Interface()
	// before loading anything to interm config mark notation map as not editable
	app.intermConfig.IsEditable(false)

	// run on host and port mentioned inside the config
	app.url = fmt.Sprintf("%v:%v", app.intermConfig.Get("host"), app.intermConfig.Get("port"))

	return nil
}

// Create retuens a rubik.Router instance for using and grouping routes.
// It is generally used if you want to add routes under the same umbrella
// prefix of this router. In Rubik it is used to group routes by domains/
// responsibilities.
func Create(index string) Router {
	return Router{
		basePath: index,
	}
}

// Use ...
func Use(router Router) {
	app.routers = append(app.routers, router)
}

// UseRoute is like rubik.Use() but attaches your route to the index Router
func UseRoute(route Route) {
	router := Router{basePath: "/"}
	router.Add(route)
	app.routers = append(app.routers, router)
}

// rHandler ...
type rHandler struct {
	fn http.HandlerFunc
}

func (rh rHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rh.fn(w, r)
}

// UseHandlerFunc converts any http,HandlerFunc into rubik.Controller
func UseHandlerFunc(fn http.HandlerFunc) Controller {
	return func(req *Request) {
		fn(&req.Writer, req.Raw)
	}
}

// UseHandler converts any http,Handler into rubik.Controller
func UseHandler(handler http.Handler) Controller {
	return func(req *Request) {
		handler.ServeHTTP(&req.Writer, req.Raw)
	}
}

// UseIntermHandler converts any func(http,Handler) http,Handler into rubik.Controller
func UseIntermHandler(intermHandler func(http.Handler) http.Handler) Controller {
	return func(req *Request) {
		rh := rHandler{}
		rh.fn = func(w http.ResponseWriter, r *http.Request) {}
		intermHandler(rh).ServeHTTP(&req.Writer, req.Raw)
	}
}

// Redirect redirects your request to the given URL with status 302 by default.
// If you want to provide a custom status for your redirection you can do that
// by passing in a custom status like so:
//
// 		func someCtl(req *Request) {
// 			req.Redirect("https://ashishshekar.com", http.StatusTemporaryRedirect)
// 		}
func (req *Request) Redirect(url string, customStatus ...int) {
	redirectStatus := http.StatusFound
	if len(customStatus) > 0 {
		redirectStatus = customStatus[0]
	}

	http.Redirect(&req.Writer, req.Raw, url, redirectStatus)
}

// Proxy does not redirect your current resource locator but
// makes an internal GET call to the specified URL to serve
// it's response as your own
func Proxy(url string) Controller {
	return func(req *Request) {
		cl := NewClient(url, time.Second*30)

		en := BlankRequestEntity{}
		en.PointTo = "@"
		en.request = req.Raw
		resp, err := cl.Get(en)
		if err != nil {
			req.Throw(500, err)
			return
		}

		req.Respond(resp.StringBody)
	}
}

// SetNotFoundHandler sets custom 404 handler
func SetNotFoundHandler(h http.Handler) {
	app.mux.NotFound = h
}

// Run will make sure all dependencies are met, resolves config and it's conflicts with
// respect to the RUBIC_ENV passed while executing. It boots all your blocks, middlewares
// message passing channels and port resolution; before starting the server.
// If this method does not find PORT that is passed as the first argument or the
// config/*RUBIC_ENV.toml then it startes at :8000.
func Run() error {
	var err error
	v, err := strconv.ParseFloat(Version, 32)
	if v > 1.0 {
		runRepl()
		return nil
	}

	env := os.Getenv("RUBIK_ENV")
	// if you are in extentions mode run only extensions and exit
	// do not run the server
	if env != "" && strings.ToLower(env) == "plugin" {
		err = boot(false, true)
		if err != nil {
			return err
		}
		return nil
	}

	err = boot(false, false)
	if err != nil {
		return err
	}

	// load port from environ
	confPort := app.intermConfig.Get("port")
	confHost := app.intermConfig.Get("host")
	if confPort == nil || confHost == nil {
		msg := "port and host must be defined inside config/default.toml or ${env}.toml"
		return errors.New(msg)
	}

	var tomlUsed string
	if env == "" || env == "development" {
		tomlUsed = "default"
	} else {
		tomlUsed = env
	}
	fmt.Println("\n\nStarted development server on: " + app.url)
	fmt.Printf("Rubik version %s, configured from \"%s.toml\"\n", Version, tomlUsed)

	return http.ListenAndServe(app.url, app.mux)
}

// Respond is a terminal function for rubik controller that sends byte response
// it wraps around your arguments for better reading
func (req *Request) Respond(data interface{}, ofType ...ByteType) {
	ty := defByteType(ofType)

	switch ty {
	case Type.HTML:
		s, ok := data.(string)
		if !ok {
			req.Throw(500, E("Error: cannot be written as HTML"))
			return
		}
		writeResponse(&req.Writer, 200, Content.HTML, []byte(s))
		break
	case Type.Text:
		s, ok := data.(string)
		if !ok {
			req.Throw(500, E("Error: cannot be written as Text"))
			return
		}
		writeResponse(&req.Writer, 200, Content.Text, []byte(s))
		break
	case Type.JSON:
		req.Writer.Header().Add(Content.Header, Content.JSON)
		req.Writer.WriteHeader(200)
		err := json.NewEncoder(&req.Writer).Encode(data)
		if err != nil {
			req.Throw(500, err)
		}
	}
}

// Throw writes an error with given status code as response
// The ByteType parameter is optional as you can convert your
// error into a JSON or plain text
//
// If you dont have an error object with you in the moment
// you can use rubik.E() to quickly wrap your string into an error
// and pass it inside this function
func (req *Request) Throw(status int, err error, btype ...ByteType) {
	ty := defByteType(btype)
	switch ty {
	case Type.Text:
		writeResponse(&req.Writer, status, Content.Text, []byte(err.Error()))
		break
	case Type.JSON:
		req.Writer.Header().Add(Content.Header, Content.JSON)
		req.Writer.WriteHeader(status)
		jsonErr := RestErrorMixin{status, err.Error()}
		json.NewEncoder(&req.Writer).Encode(&jsonErr)
		break
	}
}

// E wraps the message into an error interface and returns it. This method can be used in
// your controller for throwing error response.
//
// NOTE: this error is not stdlib errors package
// this is pkg/errors error wrapper
func E(msg string) error {
	return errors.New(msg)
}

// Ctls adds the controllers one in the order of parameters passed and
// feeds them to the bootloader
func Ctls(ctls ...Controller) []Controller {
	return ctls
}

func defByteType(typs []ByteType) ByteType {
	if len(typs) > 0 {
		return typs[0]
	}
	return Type.Text
}

func runRepl() {
	mode := os.Getenv("RUBIK_MODE")
	if mode != "" && mode == "repl" {
		err := boot(true, false)
		if err != nil {
			pkg.ErrorMsg("Error while booting: " + err.Error())
		}

		// do not run repl if it is not a rubik project
		// it is a rubik project if the pwd contains rubik.toml
		projPath := pkg.GetRubikConfigPath()
		if _, err := os.Stat(projPath); os.IsNotExist(err) {
			pkg.ErrorMsg("Not a rubik project!")
		}

		repl()
	}
}
