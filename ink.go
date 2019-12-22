package ink

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Client is a basic building block of ink http client
type Client struct {
	httpClient  http.Client
	url         string
	async       chan Response
	jwtSecret   string
	basicSecret string
}

// Response is a struct that is returned by every client after
// requrest is made successfull
type Response struct {
	Status     int
	Body       interface{}
	Raw        *http.Response
	ParsedBody interface{}
	StringBody string
}

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

// Request holds the data between the intermediate state of Client
// and PostProcessor
type Request struct {
	client       http.Client
	base         string
	path         string
	requestType  string
	json         bool
	headers      map[string]interface{}
	params       []string
	body         map[string]interface{}
	query        map[string]interface{}
	responseType reflect.Type
}

// RequestProcessor is a post processing struct that helps you infer types
type RequestProcessor struct {
	requset        *Request
	processorError error
}

// New creates a new instance of ink client
func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		url: baseURL,
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}

// Get func is used to do a GET apicall
func (c *Client) Get(path string, params ...string) *RequestProcessor {
	var req = Request{}
	req.client = c.httpClient
	req.base = c.url
	req.path = path
	req.requestType = GET
	req.params = params
	return &RequestProcessor{
		requset: &req,
	}
}

// Headers ..
func (rp *RequestProcessor) Headers(data map[string]interface{}) *RequestProcessor {
	rp.requset.headers = data
	return rp
}

// Query ..
func (rp *RequestProcessor) Query(data map[string]interface{}) *RequestProcessor {
	rp.requset.query = data
	return rp
}

// QueryOfType ..
func (rp *RequestProcessor) QueryOfType(data interface{}) *RequestProcessor {
	rp.requset.query = extractFromType(data)
	return rp
}

// Body ..
func (rp *RequestProcessor) Body(data map[string]interface{}) *RequestProcessor {
	rp.requset.body = data
	return rp
}

// BodyOfType ..
func (rp *RequestProcessor) BodyOfType(data interface{}) *RequestProcessor {
	rp.requset.body = extractFromType(data)
	return rp
}

// IsJSON ..
func (rp *RequestProcessor) IsJSON(status bool) *RequestProcessor {
	rp.requset.json = status
	return rp
}

// Infer ..
func (rp *RequestProcessor) Infer(t reflect.Type) *RequestProcessor {
	rp.requset.responseType = t
	return rp
}

// Call ..
func (rp *RequestProcessor) Call() (Response, error) {
	switch rp.requset.requestType {
	case GET:
		return doGetCall(rp.requset)
	default:
		return Response{}, errors.New("InkWrongCallError: This type of call is not allowed/supported")
	}
}

func doGetCall(req *Request) (Response, error) {
	var pathWithParams = req.path

	if !strings.HasPrefix(pathWithParams, "/") {
		pathWithParams = "/" + pathWithParams
	}

	pathWithParams, err := substituteParam(pathWithParams, req.params)
	if err != nil {
		return Response{}, err
	}

	var fullURL = req.base + pathWithParams
	fmt.Println(fullURL)

	var requestBody []byte
	var httpRequest *http.Request

	if req.json && req.body != nil {
		requestBody, err = json.Marshal(req.body)
		httpRequest, err = http.NewRequest(req.requestType, fullURL, bytes.NewBuffer(requestBody))
		httpRequest.Header.Set("Content-Type", "application/json")
	} else {
		httpRequest, err = http.NewRequest(req.requestType, fullURL, nil)
	}

	if err != nil {
		return Response{}, err
	}

	httpRequest.Header.Set(HTTP_USER_AGENT, InkClientAgent)

	resp, err := req.client.Do(httpRequest)

	if err != nil {
		return Response{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	// this means that you want ink to infer something
	if req.responseType != nil {
		var responseObject interface{}
		err = json.Unmarshal(body, &responseObject)

		if err != nil {
			return Response{}, err
		}

		if reflect.TypeOf(responseObject) != req.responseType {
			return Response{
					Status:     resp.StatusCode,
					Raw:        resp,
					StringBody: string(body),
				},
				errors.New("InkInferenceError: cannot infer the current response body to given inference type:" + req.responseType.Kind().String() + ". Take a look at Response.StringBody for proper inference.")
		}
		return Response{
			Status:     resp.StatusCode,
			Raw:        resp,
			ParsedBody: responseObject,
			StringBody: string(body),
		}, nil
	}

	return Response{
		Status:     resp.StatusCode,
		Raw:        resp,
		ParsedBody: string(body),
		StringBody: string(body),
	}, nil
}
