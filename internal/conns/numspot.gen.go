// Package conns provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package conns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

// CreateKeyPairJSONBody defines parameters for CreateKeyPair.
type CreateKeyPairJSONBody struct {
	Name string `json:"name"`
}

// ImportKeyPairJSONBody defines parameters for ImportKeyPair.
type ImportKeyPairJSONBody struct {
	Name      string `json:"name"`
	PublicKey string `json:"publicKey"`
}

// CreateKeyPairJSONRequestBody defines body for CreateKeyPair for application/json ContentType.
type CreateKeyPairJSONRequestBody CreateKeyPairJSONBody

// ImportKeyPairJSONRequestBody defines body for ImportKeyPair for application/json ContentType.
type ImportKeyPairJSONRequestBody ImportKeyPairJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// GetKeyPairs request
	GetKeyPairs(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateKeyPairWithBody request with any body
	CreateKeyPairWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateKeyPair(ctx context.Context, body CreateKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ImportKeyPairWithBody request with any body
	ImportKeyPairWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ImportKeyPair(ctx context.Context, body ImportKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteKeyPair request
	DeleteKeyPair(ctx context.Context, keyPairName string, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) GetKeyPairs(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetKeyPairsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateKeyPairWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateKeyPairRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateKeyPair(ctx context.Context, body CreateKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateKeyPairRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ImportKeyPairWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewImportKeyPairRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ImportKeyPair(ctx context.Context, body ImportKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewImportKeyPairRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteKeyPair(ctx context.Context, keyPairName string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteKeyPairRequest(c.Server, keyPairName)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewGetKeyPairsRequest generates requests for GetKeyPairs
func NewGetKeyPairsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/keyPairs")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateKeyPairRequest calls the generic CreateKeyPair builder with application/json body
func NewCreateKeyPairRequest(server string, body CreateKeyPairJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateKeyPairRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateKeyPairRequestWithBody generates requests for CreateKeyPair with any type of body
func NewCreateKeyPairRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/keyPairs")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewImportKeyPairRequest calls the generic ImportKeyPair builder with application/json body
func NewImportKeyPairRequest(server string, body ImportKeyPairJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewImportKeyPairRequestWithBody(server, "application/json", bodyReader)
}

// NewImportKeyPairRequestWithBody generates requests for ImportKeyPair with any type of body
func NewImportKeyPairRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/keyPairs/import")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteKeyPairRequest generates requests for DeleteKeyPair
func NewDeleteKeyPairRequest(server string, keyPairName string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "keyPairName", runtime.ParamLocationPath, keyPairName)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/keyPairs/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// GetKeyPairsWithResponse request
	GetKeyPairsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetKeyPairsResponse, error)

	// CreateKeyPairWithBodyWithResponse request with any body
	CreateKeyPairWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateKeyPairResponse, error)

	CreateKeyPairWithResponse(ctx context.Context, body CreateKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateKeyPairResponse, error)

	// ImportKeyPairWithBodyWithResponse request with any body
	ImportKeyPairWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ImportKeyPairResponse, error)

	ImportKeyPairWithResponse(ctx context.Context, body ImportKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*ImportKeyPairResponse, error)

	// DeleteKeyPairWithResponse request
	DeleteKeyPairWithResponse(ctx context.Context, keyPairName string, reqEditors ...RequestEditorFn) (*DeleteKeyPairResponse, error)
}

type GetKeyPairsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Items *[]struct {
			Fingerprint *string `json:"fingerprint,omitempty"`
			Name        *string `json:"name,omitempty"`
		} `json:"items,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetKeyPairsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetKeyPairsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateKeyPairResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *struct {
		Name       *string `json:"name,omitempty"`
		PrivateKey *string `json:"privateKey,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r CreateKeyPairResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateKeyPairResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ImportKeyPairResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *struct {
		Name *string `json:"name,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ImportKeyPairResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ImportKeyPairResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteKeyPairResponse struct {
	Body         []byte
	HTTPResponse *http.Response
}

// Status returns HTTPResponse.Status
func (r DeleteKeyPairResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteKeyPairResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// GetKeyPairsWithResponse request returning *GetKeyPairsResponse
func (c *ClientWithResponses) GetKeyPairsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetKeyPairsResponse, error) {
	rsp, err := c.GetKeyPairs(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetKeyPairsResponse(rsp)
}

// CreateKeyPairWithBodyWithResponse request with arbitrary body returning *CreateKeyPairResponse
func (c *ClientWithResponses) CreateKeyPairWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateKeyPairResponse, error) {
	rsp, err := c.CreateKeyPairWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateKeyPairResponse(rsp)
}

func (c *ClientWithResponses) CreateKeyPairWithResponse(ctx context.Context, body CreateKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateKeyPairResponse, error) {
	rsp, err := c.CreateKeyPair(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateKeyPairResponse(rsp)
}

// ImportKeyPairWithBodyWithResponse request with arbitrary body returning *ImportKeyPairResponse
func (c *ClientWithResponses) ImportKeyPairWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ImportKeyPairResponse, error) {
	rsp, err := c.ImportKeyPairWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseImportKeyPairResponse(rsp)
}

func (c *ClientWithResponses) ImportKeyPairWithResponse(ctx context.Context, body ImportKeyPairJSONRequestBody, reqEditors ...RequestEditorFn) (*ImportKeyPairResponse, error) {
	rsp, err := c.ImportKeyPair(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseImportKeyPairResponse(rsp)
}

// DeleteKeyPairWithResponse request returning *DeleteKeyPairResponse
func (c *ClientWithResponses) DeleteKeyPairWithResponse(ctx context.Context, keyPairName string, reqEditors ...RequestEditorFn) (*DeleteKeyPairResponse, error) {
	rsp, err := c.DeleteKeyPair(ctx, keyPairName, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteKeyPairResponse(rsp)
}

// ParseGetKeyPairsResponse parses an HTTP response from a GetKeyPairsWithResponse call
func ParseGetKeyPairsResponse(rsp *http.Response) (*GetKeyPairsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetKeyPairsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Items *[]struct {
				Fingerprint *string `json:"fingerprint,omitempty"`
				Name        *string `json:"name,omitempty"`
			} `json:"items,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateKeyPairResponse parses an HTTP response from a CreateKeyPairWithResponse call
func ParseCreateKeyPairResponse(rsp *http.Response) (*CreateKeyPairResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateKeyPairResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest struct {
			Name       *string `json:"name,omitempty"`
			PrivateKey *string `json:"privateKey,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	}

	return response, nil
}

// ParseImportKeyPairResponse parses an HTTP response from a ImportKeyPairWithResponse call
func ParseImportKeyPairResponse(rsp *http.Response) (*ImportKeyPairResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ImportKeyPairResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest struct {
			Name *string `json:"name,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	}

	return response, nil
}

// ParseDeleteKeyPairResponse parses an HTTP response from a DeleteKeyPairWithResponse call
func ParseDeleteKeyPairResponse(rsp *http.Response) (*DeleteKeyPairResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteKeyPairResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	return response, nil
}
