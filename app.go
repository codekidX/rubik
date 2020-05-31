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
	msgRegistry: make(map[string]rx),
	comm:        make(map[string]Communicator),
	routeTree: RouteTree{
		RouterList: make(map[string]string),
		Routes:     []RouteInfo{},
	},
}

var blocks = make(map[string]interface{})
var beforeHooks []RequestHook
var afterHooks []RequestHook

// Dispatch is topic dispatcher of rubik server
var Dispatch = MessagePasser{
	Message: make(chan Message),
	Error:   make(chan error),
}

const (
	// Version of rubik
	Version = "0.1"
)

type tracer interface {
	StackTrace() errors.StackTrace
}

// RestErrorMixin type is used by rubik to show error in a same format
type RestErrorMixin struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (re RestErrorMixin) Error() string {
	return re.Message
}

// Controller ...
type Controller func(*Request)

// Request ...
type Request struct {
	app    *rubik
	Entity interface{}
	Writer RResponseWriter
	Params httprouter.Params
	Raw    *http.Request
	Ctx    map[string]interface{}
}

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
	rootConfig   *pkg.Config
	logger       *pkg.Logger
	currentEnv   string
	url          string
	mux          *httprouter.Router
	blocks       map[string]Block
	afterBlocks  map[string]Block
	routers      []Router
	routeTree    RouteTree
	comm         map[string]Communicator
	msgRegistry  map[string]rx
}

// Request ...
// type RequestP struct {
// 	Raw            *http.Request
// 	Params         httprouter.Params
// 	ResponseHeader http.Header
// 	entity         interface{}
// }

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

// Route is the culmination of
type Route struct {
	Path                 string
	Method               string
	Description          string
	ResponseDeclarations map[int]string
	JSON                 bool
	Entity               interface{}
	Guard                AuthorizationGuard
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

// RouteInfo ...
type RouteInfo struct {
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

	var defaultMap map[string]interface{}
	var envMap map[string]interface{}
	var envConfigPath string

	env := os.Getenv("RUBIK_ENV")
	// set the current env to app.currentEnv
	app.currentEnv = env

	defaultConfigPath := filepath.Join(".", "config", "default.toml")
	envConfigFound := false

	if env != "" {
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

	port, _ := app.intermConfig.Get("port").(string)
	// TODO: think about this line, how do we know if we want it to
	// run on machine ip or on the localhost?
	app.url = "127.0.0.1" + port

	return nil
}

// Create ...
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

// Redirect redirects your request to the given URL
func Redirect(url string) ByteResponse {
	return ByteResponse{
		redirectURL: url,
	}
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

// SetNotFoundHandler sets custom handler for 404
func SetNotFoundHandler(h http.Handler) {
	app.mux.NotFound = h
}

// Tx is rubik's transmission method from which is intended to
// transmit the information from one place to another depending upon
// the information specified inside the target paramater
func Tx(blockName string, target string, body interface{}) {
	if app.comm[blockName] == nil {
		pkg.ErrorMsg(fmt.Sprintf("No such trasmitor %s", blockName))
		return
	}
	go app.comm[blockName].Send(target, body)
}

// Rx is recieving method of rubik which register's recievers for
// a specific topic and and handles the execution of the reciever
// on evaluating a message from the source using the controller
// passed as the parameter
func Rx(blockName string, topic string, entity interface{}, fn Controller) {
	if entity != nil {
		if reflect.ValueOf(entity).Kind() != reflect.Ptr {
			msg := fmt.Sprintf("Entity for Rx() method for topic %s requires a pointer type",
				topic)
			panic(errors.New(msg))
		}
	}

	var name = blockName
	if blockName == "" {
		name = "int"
	}
	tag := T(name, topic)
	app.msgRegistry[tag] = rx{fn, entity}
}

// Run rubik server instance
func Run(args ...string) error {
	v, err := strconv.ParseFloat(Version, 32)
	if v > 1.0 {
		runRepl()
		return nil
	}

	err = boot(false)
	if err != nil {
		return err
	}

	var port string
	if app.config != nil {
		// load port from environ
		val := app.intermConfig.Get("port")
		portVal, ok := val.(string)
		if val == nil || !ok {
			port = ":8000"
		} else {
			port = portVal
		}

	} else if app.config == nil && len(args) == 0 {
		port = ":8000"
	}

	if port == "" && len(args) > 0 {
		port = args[0]
	}

	pkg.RubikMsg("Rubik server started on port " + port[1:])

	return http.ListenAndServe(port, app.mux)
}

// Respond is a terminal function for rubik controller that sends byte response
// it wraps around your arguments for better reading
func (req *Request) Respond(data interface{}, ofType ...ByteType) {
	var ty ByteType
	if len(ofType) == 0 {
		ty = Type.Text
	} else {
		ty = ofType[0]
	}

	switch ty {
	case Type.HTML:
		s, ok := data.(string)
		if !ok {
			req.Throw(500, E("Error: cannot be written as string"))
			return
		}
		writeResponse(&req.Writer, 200, Content.HTML, []byte(s))
		break
	case Type.Text:
		s, ok := data.(string)
		if !ok {
			req.Throw(500, E("Error: cannot be written as string"))
			return
		}
		writeResponse(&req.Writer, 200, Content.Text, []byte(s))
		break
	case Type.JSON:
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
// NOTE: if you dont have an error object with you in the moment
// you can use r.E() to quickly wrap your stirng into an error
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

// T returns the tag for your receiver when you describe it in Rx() method
// Tag is always defined as
//
// BlockName : Topic string
// If BlockName is empty string then the tag uses "int" (internal)
// Denoting that the message is used for internal server purposes
func T(block string, topic string) string {
	if block == "" {
		return fmt.Sprintf("int:%s", topic)
	}
	return fmt.Sprintf("%s:%s", block, topic)
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
		err := boot(true)
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
