package rubik

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"
)

// Client is the implementation for rubik project to create
// a common abstraction of HTTP calls by passing defined entity
type Client struct {
	httpClient  http.Client
	url         string
	Debug       bool
	JWTSecret   string
	BasicSecret string
	BearerName  string
	UserAgent   string
}

// Response is a struct that is returned by every client after
// request is made successful
type Response struct {
	Status     int
	Body       interface{}
	Raw        *http.Response
	ParsedBody interface{}
	StringBody string
	IsJSON     bool
}

const (
	// GET method
	GET = "GET"
	// POST method
	POST = "POST"
	// PUT method
	PUT = "PUT"
	// DELETE method
	DELETE = "DELETE"
)

// Payload holds the data between the intermediate state of Client
// and PostProcessor
type Payload struct {
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
	rawBody      []byte
	responseType interface{}
	cancel       context.CancelFunc
	context      context.Context
	agent        string
}

// NewClient creates a new instance of rubik client
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
	if err != nil {
		return Response{}, err
	}
	req.requestType = GET
	return call(req)
}

// Post ...
func (c *Client) Post(entity interface{}) (Response, error) {
	req, err := populateRequest(entity, c)
	if err != nil {
		return Response{}, err
	}
	req.requestType = POST
	return call(req)
}

// Put ...
func (c *Client) Put(entity interface{}) (Response, error) {
	req, err := populateRequest(entity, c)
	if err != nil {
		return Response{}, err
	}
	req.requestType = PUT
	return call(req)
}

// Delete ...
func (c *Client) Delete(entity interface{}) (Response, error) {
	req, err := populateRequest(entity, c)
	if err != nil {
		return Response{}, err
	}
	req.requestType = DELETE
	return call(req)
}

// Download method downloads file from an url from your specified Entity->Route
// to TargetFilePath passed to the entity
func (c *Client) Download(entity DownloadRequestEntity) ([]byte, error) {
	if entity.PointTo == "" {
		errMsg := "DownloadRequestEntity must have a route initialized using Route() method"
		return nil, errors.New(errMsg)
	}

	// source
	finalURL := c.url + safeRoutePath(entity.PointTo)
	raw, err := downloadCall(finalURL, entity.TargetFilePath)

	if err != nil {
		return nil, err
	}

	return raw, nil
}

// Cancel ...
func (r *Payload) Cancel() {
	r.cancel()
}

func downloadCall(url, target string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, errors.New("DownloadError: Cannot download file. Raw: " + err.Error())
	}
	defer resp.Body.Close()

	out, err := os.Create(target)

	if err != nil {
		return nil, errors.New("DownloadError: Cannot create target file. Raw: " + err.Error())
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, errors.New("DownloadError: Cannot copy to target file. Raw: " + err.Error())
	}

	b, err := ioutil.ReadFile(target)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func call(req *Payload) (Response, error) {

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
			errMsg := "InferenceError: Cannot infer a non-json/non-mappable value to " +
				"specified type: %s. Response.ParsedBody/Response.StringBody of type map is " +
				"present for access."
			message := fmt.Sprintf(errMsg, responseType.Name())
			return Response{
				Status:     resp.StatusCode,
				IsJSON:     false,
				StringBody: string(body),
			}, errors.New(message)
		}

		return Response{
			Status:     resp.StatusCode,
			Raw:        resp,
			IsJSON:     true,
			ParsedBody: req.responseType,
			StringBody: string(body),
		}, nil
	}

	return Response{
		Status:     resp.StatusCode,
		Raw:        resp,
		IsJSON:     false,
		ParsedBody: req.responseType,
		StringBody: string(body),
	}, nil
}
