// Package iam provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.0.0 DO NOT EDIT.
package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

const (
	BasicScopes  = "basic.Scopes"
	Oauth2Scopes = "oauth2.Scopes"
)

// ErrorOAuth2 Error
type ErrorOAuth2 struct {
	// Error Error
	Error *string `json:"error,omitempty"`

	// ErrorDebug Error Debug Information
	//
	// Only available in dev mode.
	ErrorDebug *string `json:"error_debug,omitempty"`

	// ErrorDescription Error Description
	ErrorDescription *string `json:"error_description,omitempty"`

	// ErrorHint Error Hint
	//
	// Helps the user identify the error cause.
	ErrorHint *string `json:"error_hint,omitempty"`

	// StatusCode HTTP Status Code
	StatusCode *int64 `json:"status_code,omitempty"`
}

// IntrospectedOAuth2Token Introspection contains an access token's session data as specified by
// [IETF RFC 7662](https://tools.ietf.org/html/rfc7662)
type IntrospectedOAuth2Token struct {
	// Active Active is a boolean indicator of whether or not the presented token
	// is currently active.  The specifics of a token's "active" state
	// will vary depending on the implementation of the authorization
	// server and the information it keeps about its tokens, but a "true"
	// value return for the "active" property will generally indicate
	// that a given token has been issued by this authorization server,
	// has not been revoked by the resource owner, and is within its
	// given time window of validity (e.g., after its issuance time and
	// before its expiration time).
	Active bool `json:"active"`

	// Aud Audience contains a list of the token's intended audiences.
	Aud *[]string `json:"aud,omitempty"`

	// ClientId ID is a client identifier for the OAuth 2.0 client that
	// requested this token.
	ClientId *string `json:"client_id,omitempty"`

	// Exp Expires at is an integer timestamp, measured in the number of seconds
	// since January 1 1970 UTC, indicating when this token will expire.
	Exp *int64 `json:"exp,omitempty"`

	// Iat Issued at is an integer timestamp, measured in the number of seconds
	// since January 1 1970 UTC, indicating when this token was
	// originally issued.
	Iat *int64 `json:"iat,omitempty"`

	// Iss IssuerURL is a string representing the issuer of this token
	Iss *string `json:"iss,omitempty"`

	// Nbf NotBefore is an integer timestamp, measured in the number of seconds
	// since January 1 1970 UTC, indicating when this token is not to be
	// used before.
	Nbf *int64 `json:"nbf,omitempty"`

	// Scope Scope is a JSON string containing a space-separated list of
	// scopes associated with this token.
	Scope *string `json:"scope,omitempty"`

	// Sub Subject of the token, as defined in JWT [RFC7519].
	// Usually a machine-readable identifier of the resource owner who
	// authorized this token.
	Sub *string `json:"sub,omitempty"`

	// TokenType TokenType is the introspected token's type, typically `Bearer`.
	TokenType *string `json:"token_type,omitempty"`

	// TokenUse TokenUse is the introspected token's use, for example `access_token` or `refresh_token`.
	TokenUse *string `json:"token_use,omitempty"`
}

// OAuth2TokenExchange OAuth2 Token Exchange Result
type OAuth2TokenExchange struct {
	// AccessToken The access token issued by the authorization server.
	AccessToken *string `json:"access_token,omitempty"`

	// ExpiresIn The lifetime in seconds of the access token. For
	// example, the value "3600" denotes that the access token will
	// expire in one hour from the time the response was generated.
	ExpiresIn *int64 `json:"expires_in,omitempty"`

	// IdToken To retrieve a refresh token request the id_token scope.
	IdToken *string `json:"id_token,omitempty"`

	// RefreshToken The refresh token, which can be used to obtain new
	// access tokens. To retrieve it add the scope "offline" to your access token request.
	RefreshToken *string `json:"refresh_token,omitempty"`

	// Scope The scope of the access token
	Scope *string `json:"scope,omitempty"`

	// TokenType The type of the token issued
	TokenType *string `json:"token_type,omitempty"`
}

// IntrospectOAuth2TokenFormdataBody defines parameters for IntrospectOAuth2Token.
type IntrospectOAuth2TokenFormdataBody struct {
	// Scope An optional, space separated list of required scopes. If the access token was not granted one of the
	// scopes, the result of active will be false.
	Scope *string `form:"scope,omitempty" json:"scope,omitempty"`

	// Token The string value of the token. For access tokens, this
	// is the "access_token" value returned from the token endpoint
	// defined in OAuth 2.0. For refresh tokens, this is the "refresh_token"
	// value returned.
	Token string `form:"token" json:"token"`
}

// RevokeOAuth2TokenFormdataBody defines parameters for RevokeOAuth2Token.
type RevokeOAuth2TokenFormdataBody struct {
	ClientId     *string `form:"client_id,omitempty" json:"client_id,omitempty"`
	ClientSecret *string `form:"client_secret,omitempty" json:"client_secret,omitempty"`
	Token        string  `form:"token" json:"token"`
}

// Oauth2TokenExchangeFormdataBody defines parameters for Oauth2TokenExchange.
type Oauth2TokenExchangeFormdataBody struct {
	ClientId     *string `form:"client_id,omitempty" json:"client_id,omitempty"`
	Code         *string `form:"code,omitempty" json:"code,omitempty"`
	GrantType    string  `form:"grant_type" json:"grant_type"`
	RedirectUri  *string `form:"redirect_uri,omitempty" json:"redirect_uri,omitempty"`
	RefreshToken *string `form:"refresh_token,omitempty" json:"refresh_token,omitempty"`
}

// IntrospectOAuth2TokenFormdataRequestBody defines body for IntrospectOAuth2Token for application/x-www-form-urlencoded ContentType.
type IntrospectOAuth2TokenFormdataRequestBody IntrospectOAuth2TokenFormdataBody

// RevokeOAuth2TokenFormdataRequestBody defines body for RevokeOAuth2Token for application/x-www-form-urlencoded ContentType.
type RevokeOAuth2TokenFormdataRequestBody RevokeOAuth2TokenFormdataBody

// Oauth2TokenExchangeFormdataRequestBody defines body for Oauth2TokenExchange for application/x-www-form-urlencoded ContentType.
type Oauth2TokenExchangeFormdataRequestBody Oauth2TokenExchangeFormdataBody

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
	// OAuth2Authorize request
	OAuth2Authorize(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// IntrospectOAuth2TokenWithBody request with any body
	IntrospectOAuth2TokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	IntrospectOAuth2TokenWithFormdataBody(ctx context.Context, body IntrospectOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RevokeOAuth2TokenWithBody request with any body
	RevokeOAuth2TokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	RevokeOAuth2TokenWithFormdataBody(ctx context.Context, body RevokeOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Oauth2TokenExchangeWithBody request with any body
	Oauth2TokenExchangeWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	Oauth2TokenExchangeWithFormdataBody(ctx context.Context, body Oauth2TokenExchangeFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) OAuth2Authorize(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewOAuth2AuthorizeRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) IntrospectOAuth2TokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewIntrospectOAuth2TokenRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) IntrospectOAuth2TokenWithFormdataBody(ctx context.Context, body IntrospectOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewIntrospectOAuth2TokenRequestWithFormdataBody(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RevokeOAuth2TokenWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRevokeOAuth2TokenRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RevokeOAuth2TokenWithFormdataBody(ctx context.Context, body RevokeOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRevokeOAuth2TokenRequestWithFormdataBody(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Oauth2TokenExchangeWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewOauth2TokenExchangeRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Oauth2TokenExchangeWithFormdataBody(ctx context.Context, body Oauth2TokenExchangeFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewOauth2TokenExchangeRequestWithFormdataBody(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewOAuth2AuthorizeRequest generates requests for OAuth2Authorize
func NewOAuth2AuthorizeRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/authorize")
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

// NewIntrospectOAuth2TokenRequestWithFormdataBody calls the generic IntrospectOAuth2Token builder with application/x-www-form-urlencoded body
func NewIntrospectOAuth2TokenRequestWithFormdataBody(server string, body IntrospectOAuth2TokenFormdataRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	bodyStr, err := runtime.MarshalForm(body, nil)
	if err != nil {
		return nil, err
	}
	bodyReader = strings.NewReader(bodyStr.Encode())
	return NewIntrospectOAuth2TokenRequestWithBody(server, "application/x-www-form-urlencoded", bodyReader)
}

// NewIntrospectOAuth2TokenRequestWithBody generates requests for IntrospectOAuth2Token with any type of body
func NewIntrospectOAuth2TokenRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/introspect")
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

// NewRevokeOAuth2TokenRequestWithFormdataBody calls the generic RevokeOAuth2Token builder with application/x-www-form-urlencoded body
func NewRevokeOAuth2TokenRequestWithFormdataBody(server string, body RevokeOAuth2TokenFormdataRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	bodyStr, err := runtime.MarshalForm(body, nil)
	if err != nil {
		return nil, err
	}
	bodyReader = strings.NewReader(bodyStr.Encode())
	return NewRevokeOAuth2TokenRequestWithBody(server, "application/x-www-form-urlencoded", bodyReader)
}

// NewRevokeOAuth2TokenRequestWithBody generates requests for RevokeOAuth2Token with any type of body
func NewRevokeOAuth2TokenRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/revoke")
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

// NewOauth2TokenExchangeRequestWithFormdataBody calls the generic Oauth2TokenExchange builder with application/x-www-form-urlencoded body
func NewOauth2TokenExchangeRequestWithFormdataBody(server string, body Oauth2TokenExchangeFormdataRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	bodyStr, err := runtime.MarshalForm(body, nil)
	if err != nil {
		return nil, err
	}
	bodyReader = strings.NewReader(bodyStr.Encode())
	return NewOauth2TokenExchangeRequestWithBody(server, "application/x-www-form-urlencoded", bodyReader)
}

// NewOauth2TokenExchangeRequestWithBody generates requests for Oauth2TokenExchange with any type of body
func NewOauth2TokenExchangeRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/token")
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
	// OAuth2AuthorizeWithResponse request
	OAuth2AuthorizeWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*OAuth2AuthorizeResponse, error)

	// IntrospectOAuth2TokenWithBodyWithResponse request with any body
	IntrospectOAuth2TokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*IntrospectOAuth2TokenResponse, error)

	IntrospectOAuth2TokenWithFormdataBodyWithResponse(ctx context.Context, body IntrospectOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*IntrospectOAuth2TokenResponse, error)

	// RevokeOAuth2TokenWithBodyWithResponse request with any body
	RevokeOAuth2TokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RevokeOAuth2TokenResponse, error)

	RevokeOAuth2TokenWithFormdataBodyWithResponse(ctx context.Context, body RevokeOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*RevokeOAuth2TokenResponse, error)

	// Oauth2TokenExchangeWithBodyWithResponse request with any body
	Oauth2TokenExchangeWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*Oauth2TokenExchangeResponse, error)

	Oauth2TokenExchangeWithFormdataBodyWithResponse(ctx context.Context, body Oauth2TokenExchangeFormdataRequestBody, reqEditors ...RequestEditorFn) (*Oauth2TokenExchangeResponse, error)
}

type OAuth2AuthorizeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSONDefault  *ErrorOAuth2
}

// Status returns HTTPResponse.Status
func (r OAuth2AuthorizeResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r OAuth2AuthorizeResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type IntrospectOAuth2TokenResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *IntrospectedOAuth2Token
	JSONDefault  *ErrorOAuth2
}

// Status returns HTTPResponse.Status
func (r IntrospectOAuth2TokenResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r IntrospectOAuth2TokenResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RevokeOAuth2TokenResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSONDefault  *ErrorOAuth2
}

// Status returns HTTPResponse.Status
func (r RevokeOAuth2TokenResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RevokeOAuth2TokenResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type Oauth2TokenExchangeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *OAuth2TokenExchange
	JSONDefault  *ErrorOAuth2
}

// Status returns HTTPResponse.Status
func (r Oauth2TokenExchangeResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r Oauth2TokenExchangeResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// OAuth2AuthorizeWithResponse request returning *OAuth2AuthorizeResponse
func (c *ClientWithResponses) OAuth2AuthorizeWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*OAuth2AuthorizeResponse, error) {
	rsp, err := c.OAuth2Authorize(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseOAuth2AuthorizeResponse(rsp)
}

// IntrospectOAuth2TokenWithBodyWithResponse request with arbitrary body returning *IntrospectOAuth2TokenResponse
func (c *ClientWithResponses) IntrospectOAuth2TokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*IntrospectOAuth2TokenResponse, error) {
	rsp, err := c.IntrospectOAuth2TokenWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseIntrospectOAuth2TokenResponse(rsp)
}

func (c *ClientWithResponses) IntrospectOAuth2TokenWithFormdataBodyWithResponse(ctx context.Context, body IntrospectOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*IntrospectOAuth2TokenResponse, error) {
	rsp, err := c.IntrospectOAuth2TokenWithFormdataBody(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseIntrospectOAuth2TokenResponse(rsp)
}

// RevokeOAuth2TokenWithBodyWithResponse request with arbitrary body returning *RevokeOAuth2TokenResponse
func (c *ClientWithResponses) RevokeOAuth2TokenWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RevokeOAuth2TokenResponse, error) {
	rsp, err := c.RevokeOAuth2TokenWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRevokeOAuth2TokenResponse(rsp)
}

func (c *ClientWithResponses) RevokeOAuth2TokenWithFormdataBodyWithResponse(ctx context.Context, body RevokeOAuth2TokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*RevokeOAuth2TokenResponse, error) {
	rsp, err := c.RevokeOAuth2TokenWithFormdataBody(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRevokeOAuth2TokenResponse(rsp)
}

// Oauth2TokenExchangeWithBodyWithResponse request with arbitrary body returning *Oauth2TokenExchangeResponse
func (c *ClientWithResponses) Oauth2TokenExchangeWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*Oauth2TokenExchangeResponse, error) {
	rsp, err := c.Oauth2TokenExchangeWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseOauth2TokenExchangeResponse(rsp)
}

func (c *ClientWithResponses) Oauth2TokenExchangeWithFormdataBodyWithResponse(ctx context.Context, body Oauth2TokenExchangeFormdataRequestBody, reqEditors ...RequestEditorFn) (*Oauth2TokenExchangeResponse, error) {
	rsp, err := c.Oauth2TokenExchangeWithFormdataBody(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseOauth2TokenExchangeResponse(rsp)
}

// ParseOAuth2AuthorizeResponse parses an HTTP response from a OAuth2AuthorizeWithResponse call
func ParseOAuth2AuthorizeResponse(rsp *http.Response) (*OAuth2AuthorizeResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &OAuth2AuthorizeResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest ErrorOAuth2
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseIntrospectOAuth2TokenResponse parses an HTTP response from a IntrospectOAuth2TokenWithResponse call
func ParseIntrospectOAuth2TokenResponse(rsp *http.Response) (*IntrospectOAuth2TokenResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &IntrospectOAuth2TokenResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest IntrospectedOAuth2Token
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest ErrorOAuth2
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseRevokeOAuth2TokenResponse parses an HTTP response from a RevokeOAuth2TokenWithResponse call
func ParseRevokeOAuth2TokenResponse(rsp *http.Response) (*RevokeOAuth2TokenResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RevokeOAuth2TokenResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest ErrorOAuth2
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseOauth2TokenExchangeResponse parses an HTTP response from a Oauth2TokenExchangeWithResponse call
func ParseOauth2TokenExchangeResponse(rsp *http.Response) (*Oauth2TokenExchangeResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &Oauth2TokenExchangeResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest OAuth2TokenExchange
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest ErrorOAuth2
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}
