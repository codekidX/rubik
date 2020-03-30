package rubik

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

func populateRequest(entity interface{}, c *Client) (*Payload, error) {
	isEntity := checkIsEntity(entity)

	if !isEntity {
		panic(errors.New("EntityError: The entity you are passing must extend RequestEntity. "))
	}

	req, err := extract(entity)

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

func populateParamsAndQuery(req *Payload) (string, error) {
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

func populateHTTPRequest(req *Payload, fullURL string) (*http.Request, error) {
	var err error
	var requestBody []byte
	var httpRequest *http.Request

	if req.json && len(req.body) > 0 {
		requestBody, err = json.Marshal(req.body)
		httpRequest, err = http.NewRequest(req.requestType, fullURL, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		httpRequest.Header.Set(Content.Header, Content.JSON)
	} else if req.urlencoded && len(req.body) > 0 {
		requestBody = []byte(req.body.Encode())
		httpRequest, err = http.NewRequest(req.requestType, fullURL, bytes.NewBuffer(requestBody))
		httpRequest.Header.Set(Content.Header, Content.URLEncoded)
	} else if req.formData && req.formBody.Len() > 0 {
		httpRequest, err = http.NewRequest(req.requestType, fullURL, req.formBody)
		// httpRequest.Header.Set("Content-Type", "multipart/form-data")
	} else {
		httpRequest, err = http.NewRequest(req.requestType, fullURL, nil)
	}

	if len(req.headers) > 0 {
		for k, v := range req.headers {
			httpRequest.Header.Set(k, v[0])
		}
	}

	httpRequest = httpRequest.WithContext(req.context)

	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set(headerUserAgent, clientAgent)

	return httpRequest, nil
}
