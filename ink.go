package ink

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// Client is a basic building block of ink http client
type Client struct {
	httpClient  http.Client
	url         string
	async       chan Response
	JWTSecret   string
	BasicSecret string
	BearerName  string
}

// Response is a struct that is returned by every client after
// requrest is made successfull
type Response struct {
	Status     int
	Body       interface{}
	Raw        *http.Response
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
	urlencoded   bool
	headers      url.Values
	params       []string
	body         url.Values
	query        url.Values
	responseType interface{}
	cancel       context.CancelFunc
	context      context.Context
}

// RequestProcessor is a post processing struct that helps you infer types
type RequestProcessor struct {
	requset        *Request
	Headers        url.Values
	Body           url.Values
	Query          url.Values
	processorError error
}

// New creates a new instance of ink client
func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		url: baseURL,
		httpClient: http.Client{
			Timeout: timeout,
		},
		BearerName: "Bearer",
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
		Body:    url.Values{},
		Query:   url.Values{},
	}
}

// QueryOfType ..
func (rp *RequestProcessor) QueryOfType(data interface{}) *RequestProcessor {
	rp.requset.query = extractFromType(data)
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

// IsURLEncoded ..
func (rp *RequestProcessor) IsURLEncoded(status bool) *RequestProcessor {
	rp.requset.urlencoded = status
	return rp
}

// Infer ..
func (rp *RequestProcessor) Infer(target interface{}) *RequestProcessor {
	rp.requset.responseType = target
	return rp
}

// Cancel ...
func (rp *RequestProcessor) Cancel() {
	rp.requset.cancel()
}

// Call ..
func (rp *RequestProcessor) Call() (Response, error) {
	ctx, cancel := context.WithCancel(context.Background())
	rp.requset.cancel = cancel
	rp.requset.context = ctx
	rp.requset.body = rp.Body
	rp.requset.query = rp.Query

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

	if len(req.query) > 0 {
		fullURL += "?" + req.query.Encode()
	}

	var requestBody []byte
	var httpRequest *http.Request

	if req.json && req.body != nil {
		requestBody, err = json.Marshal(req.body)
		httpRequest, err = http.NewRequest(req.requestType, fullURL, bytes.NewBuffer(requestBody))
		httpRequest.Header.Set("Content-Type", "application/json")
	} else if req.urlencoded && req.body != nil {
		requestBody = []byte(req.body.Encode())
		httpRequest, err = http.NewRequest(req.requestType, fullURL, bytes.NewBuffer(requestBody))
	} else {
		httpRequest, err = http.NewRequest(req.requestType, fullURL, nil)
	}

	httpRequest = httpRequest.WithContext(req.context)

	if err != nil {
		return Response{}, err
	}

	httpRequest.Header.Set(HeaderUserAgent, InkClientAgent)

	resp, err := req.client.Do(httpRequest)

	if err != nil {
		return Response{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	responseType := reflect.TypeOf(req.responseType)
	// this means that you want ink to infer something
	if responseType != nil {
		err = json.Unmarshal(body, req.responseType)

		if err != nil {
			message := fmt.Sprintf("InkInferenceError: Cannot infer a non-json value to specified type: %s", responseType.Name())
			return Response{
				Status: resp.StatusCode,
			}, errors.New(message)
		}

		return Response{
			Status:     resp.StatusCode,
			Raw:        resp,
			StringBody: string(body),
		}, nil
	}

	return Response{
		Status:     resp.StatusCode,
		Raw:        resp,
		StringBody: string(body),
	}, nil
}
