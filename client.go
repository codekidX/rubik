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
	"time"
)

// Client is a basic building block of ink http client
type Client struct {
	httpClient  http.Client
	url         string
	async       chan Response
	Debug       bool
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
	urlencoded   bool
	formData     bool
	headers      url.Values
	params       []string
	body         Values
	query        url.Values
	formBody     *bytes.Buffer
	responseType interface{}
	cancel       context.CancelFunc
	context      context.Context
}

// NewClient creates a new instance of ink client
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		url: baseURL,
		httpClient: http.Client{
			Timeout: timeout,
		},
		BearerName: "Bearer",
	}
}

// Get ...
func (c *Client) Get(entity interface{}) (Response, error) {
	req, err := populateRequest(entity, c)
	req.requestType = GET
	if err != nil {
		return Response{}, err
	}
	return call(req)
}

// Post ...
func (c *Client) Post(entity interface{}) (Response, error) {
	req, err := populateRequest(entity, c)
	req.requestType = POST
	if err != nil {
		return Response{}, err
	}
	return call(req)
}

// Put ...
func (c *Client) Put(entity interface{}) (Response, error) {
	req, err := populateRequest(entity, c)
	req.requestType = PUT
	if err != nil {
		return Response{}, err
	}
	return call(req)
}

// Delete ...
func (c *Client) Delete(entity interface{}) (Response, error) {
	req, err := populateRequest(entity, c)
	req.requestType = DELETE
	if err != nil {
		return Response{}, err
	}
	return call(req)
}

// Cancel ...
func (r *Request) Cancel() {
	r.cancel()
}

func call(req *Request) (Response, error) {

	fullURL, err := populateParamsAndQuery(req)

	if err != nil {
		return Response{}, err
	}

	httpRequest, err := populateHTTPRequest(req, fullURL)

	if err != nil {
		return Response{}, err
	}

	resp, err := req.client.Do(httpRequest)

	if err != nil {
		return Response{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	// this means that you want ink to infer something
	if req.responseType != nil {
		responseType := reflect.TypeOf(req.responseType)

		err = json.Unmarshal(body, req.responseType)

		if err != nil {
			message := fmt.Sprintf("InkInferenceError: Cannot infer a non-json/non-mappable value to specified type: %s. Response.ParsedBody/Response.StringBody of type map is present for access.", responseType.Name())
			return Response{
				Status:     resp.StatusCode,
				StringBody: string(body),
			}, errors.New(message)
		}

		return Response{
			Status:     resp.StatusCode,
			Raw:        resp,
			ParsedBody: req.responseType,
			StringBody: string(body),
		}, nil
	}

	return Response{
		Status:     resp.StatusCode,
		Raw:        resp,
		ParsedBody: req.responseType,
		StringBody: string(body),
	}, nil
}
