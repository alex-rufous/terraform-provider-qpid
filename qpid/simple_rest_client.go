package qpid

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

// Credentials interface for encapsulation of request credentials
type Credentials interface {
	Set(request *http.Request) error
}

// BasicAuthCredentials represents basic authentication credentials
type BasicAuthCredentials struct {
	Username string
	Password string
}

// Set basic authentication credentials on http request
func (c BasicAuthCredentials) Set(request *http.Request) error {
	log.Printf("[DEBUG] Qpid: BasicAuthCredentials %s", c.Username)
	request.SetBasicAuth(c.Username, c.Password)
	return nil
}

// String ...
func (c *BasicAuthCredentials) String() string {
	return "BasicAuthCredentials [username: " + c.Username + "]"
}

// NewBasicAuthCredentials constructs BasicAuthCredentials
func NewBasicAuthCredentials(username string, password string) *Credentials {
	var me Credentials = BasicAuthCredentials{
		Username: username,
		Password: password,
	}

	return &me
}

// SimpleRestClient is a basic REST client for calling REST API using GET/POST/PUT/DELETE methods
type SimpleRestClient struct {
	endpoint    *url.URL
	credentials *Credentials
	transport   http.RoundTripper
	timeout     time.Duration
}

// NewSimpleRestClient creates a client  for given URI, credentials, and transport
func NewSimpleRestClient(uri string, credentials *Credentials, transport http.RoundTripper) (me *SimpleRestClient, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return &SimpleRestClient{}, err
	}

	me = &SimpleRestClient{
		endpoint:    u,
		credentials: credentials,
		transport:   transport,
	}

	return me, nil
}

// SetTransport sets transport
func (c *SimpleRestClient) SetTransport(transport http.RoundTripper) {
	c.transport = transport
}

// SetTimeout sets timeout
func (c *SimpleRestClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetAsMap sends GET request to the given path and returns results as map
func (c *SimpleRestClient) GetAsMap(path string, query url.Values) (*map[string]interface{}, error) {
	req, err := c.newGetHTTPRequestWithParameters(path, query)
	if err != nil {
		return &map[string]interface{}{}, err
	}

	res, err := c.executeHTTPRequest(req)
	if err != nil {
		return &map[string]interface{}{}, err
	}

	return convertHttpResponseToMap(res)
}

// Submit sends given map of attributes to the server into given path using given method
func (c *SimpleRestClient) Submit(method string, path string, attributes *map[string]interface{}) (res *http.Response, err error) {
	body, err := json.Marshal(*attributes)
	if err != nil {
		return &http.Response{}, err
	}

	req, err := c.newHTTPRequest(method, path, body)
	if err != nil {
		return &http.Response{}, err
	}

	return c.executeHTTPRequest(req)
}

// Post post attribute map into given path
func (c *SimpleRestClient) Post(path string, attributes *map[string]interface{}) (res *http.Response, err error) {
	return c.Submit(http.MethodPost, path, attributes)
}

// Put puts attribute map into given path
func (c *SimpleRestClient) Put(path string, attributes *map[string]interface{}) (res *http.Response, err error) {
	return c.Submit(http.MethodPut, path, attributes)
}

// Delete deletes the resource
func (c *SimpleRestClient) Delete(path string) (*http.Response, error) {
	req, err := c.newHTTPRequest(http.MethodDelete, path, nil)
	if err != nil {
		return &http.Response{}, err
	}

	return c.executeHTTPRequest(req)
}

func (c *SimpleRestClient) newGetHTTPRequest(path string) (*http.Request, error) {
	return c.newHTTPRequest(http.MethodGet, path, nil)
}

func (c *SimpleRestClient) newGetHTTPRequestWithParameters(path string, qs url.Values) (*http.Request, error) {
	return c.newHTTPRequest(http.MethodGet, path+"?"+qs.Encode(), nil)
}

func (c *SimpleRestClient) newHTTPRequest(method string, path string, body []byte) (*http.Request, error) {
	var b io.Reader = nil
	if body != nil {
		b = bytes.NewReader(body)
	}

	uri := c.endpoint.String() + "/" + path

	credentials := (*c).credentials
	log.Printf("new http request to %s with credentials %v", uri, credentials)
	req, err := http.NewRequest(method, uri, b)
	req.Close = true
	if err == nil {
		err = (*c.credentials).Set(req)
	}

	if err == nil && body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	return req, err
}

func (c *SimpleRestClient) executeHTTPRequest(req *http.Request) (res *http.Response, err error) {
	httpClient := &http.Client{
		Timeout: c.timeout,
	}
	if c.transport != nil {
		httpClient.Transport = c.transport
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode == 401 {
		return resp, errors.New("error: 401 unauthorized")
	}
	return resp, err
}

func (c *SimpleRestClient) GetAsArray(path string, query url.Values) (*[]map[string]interface{}, error) {
	req, err := c.newGetHTTPRequestWithParameters(path, query)
	if err != nil {
		return &[]map[string]interface{}{}, err
	}

	res, err := c.executeHTTPRequest(req)
	if err != nil {
		return &[]map[string]interface{}{}, err
	}

	return convertHttpResponseToArray(res)
}
