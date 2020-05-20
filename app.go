package rubik

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

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

// Middleware intercepts user request and processes it
type Middleware func(req Request) interface{}

// Controller ...
type Controller func(entity interface{}) ByteResponse

// RequestContext ...
type RequestContext struct {
	Request  *http.Request
	Ctx      map[string]interface{}
	Response interface{}
	Status   int
}

// RequestHook ...
type RequestHook func(*RequestContext)

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
type Request struct {
	Raw            *http.Request
	Params         httprouter.Params
	ResponseHeader http.Header
	app            *rubik
	entity         interface{}
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

// Route is the culmination of
type Route struct {
	Path                 string
	Method               string
	Description          string
	ResponseDeclarations map[int]string
	JSON                 bool
	Entity               interface{}
	Guard                AuthorizationGuard
	Middlewares          []Middleware
	Validation           Validation
	Controller           Controller
}

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

// FromStorage returns the file bytes of a given fileName as response
func FromStorage(fileName string) ByteResponse {
	var filePath = filepath.Join(".", "storage", fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return Failure(500, errors.New("FileNotFoundError: "+fileName+" does not exist."))
	}

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return Failure(500, err)
	}

	return Success(b, Type.Bytes)
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

// BeforeRequest ...
func BeforeRequest(h RequestHook) {
	beforeHooks = append(beforeHooks, h)
}

// AfterRequest ...
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

func UseRoute(route Route) {
	router := Router{basePath: "/"}
	router.Add(route)
	app.routers = append(app.routers, router)
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

// RestError returns a json with the error code and the message
func RestError(code int, message string) (interface{}, error) {
	return nil, RestErrorMixin{Code: code, Message: message}
}

// Success is a terminal function for rubik controller that sends byte response
// it wraps around your arguments for better reading
func Success(data interface{}, btype ...ByteType) ByteResponse {
	strings.Contains("a", "a")
	return ByteResponse{
		Status: http.StatusOK,
		Data:   data,
		OfType: defByteType(btype),
	}
}

// Failure returns a ByteResponse type with given status code
// The ByteType parameter is optional as you can convert your
// error into a JSON or plain text
// FACT: if you dont have an error object with you in the moment
// you can use r.E() to quickly wrap your stirng into an error
// and pass it inside this function
func Failure(status int, err error, btype ...ByteType) ByteResponse {
	return ByteResponse{
		Status: status,
		Error:  err,
		OfType: defByteType(btype),
	}
}

// E returns errors.New's error object
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
