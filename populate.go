package cherry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func populateRequest(entity interface{}, c *Client) (*Request, error) {
	isEntity := checkIsEntity(entity)

	if !isEntity {
		panic(errors.New("InkEntityError: The entity you are passing must extend RequestEntity. Take a look at codekidx.github.io/ink for more info"))
	}

	req, err := extractFromType(entity)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	req.cancel = cancel
	req.context = ctx

	req.client = c.httpClient
	req.base = c.url

	return &req, nil
}

func populateParamsAndQuery(req *Request) (string, error) {
	var pathWithParams = req.path

	if !strings.HasPrefix(pathWithParams, "/") {
		pathWithParams = "/" + pathWithParams
	}

	pathWithParams, err := substituteParam(pathWithParams, req.params)
	if err != nil {
		return "", err
	}

	var fullURL = req.base + pathWithParams

	if len(req.query) > 0 {
		fullURL += "?" + req.query.Encode()
	}

	return fullURL, nil
}

func populateHTTPRequest(req *Request, fullURL string) (*http.Request, error) {
	var err error
	var requestBody []byte
	var httpRequest *http.Request

	if req.json && len(req.body) > 0 {
		requestBody, err = json.Marshal(req.body)
		httpRequest, err = http.NewRequest(req.requestType, fullURL, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		httpRequest.Header.Set("Content-Type", "application/json")
	} else if req.urlencoded && len(req.body) > 0 {
		fmt.Println("gng url", req.body.Encode())
		requestBody = []byte(req.body.Encode())
		httpRequest, err = http.NewRequest(req.requestType, fullURL, bytes.NewBuffer(requestBody))
		httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else if req.formData && req.formBody.Len() > 0 {
		httpRequest, err = http.NewRequest(req.requestType, fullURL, req.formBody)
		// httpRequest.Header.Set("Content-Type", "multipart/form-data")
	} else {
		httpRequest, err = http.NewRequest(req.requestType, fullURL, nil)
	}

	for k, v := range req.headers {
		httpRequest.Header.Set(k , v[0])
	}

	httpRequest = httpRequest.WithContext(req.context)

	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set(HeaderUserAgent, InkClientAgent)

	return httpRequest, nil
}
