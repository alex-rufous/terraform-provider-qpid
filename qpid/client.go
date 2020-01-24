package qpid

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client Qpid REST API client
type Client struct {
	modelVersion string
	restClient   *SimpleRestClient
}

// NewClient creates a Qpid client for given URI, credentials, model version and transport
func NewClient(uri string, username string, password string, modelVersion string, transport http.RoundTripper) (me *Client, err error) {

	log.Printf("Qpid Client for endpoint: %s, model: %s, username %s", uri, modelVersion, username)

	credentials := NewBasicAuthCredentials(username, password)
	restClient, err := NewSimpleRestClient(uri+"/api/"+modelVersion, credentials, transport)
	if err != nil {
		return &Client{}, nil
	}
	me = &Client{
		restClient:   restClient,
		modelVersion: modelVersion,
	}

	return me, nil
}

// SetTransport ...
func (c *Client) SetTransport(transport http.RoundTripper) {
	c.restClient.SetTransport(transport)
}

// SetTimeout ...
func (c *Client) SetTimeout(timeout time.Duration) {
	c.restClient.SetTimeout(timeout)
}

// CreateVirtualHostNode ...
func (c *Client) CreateVirtualHostNode(attributes *map[string]interface{}) (res *http.Response, err error) {
	return c.restClient.Post("virtualhostnode", attributes)
}

// GetVirtualHostNode ...
func (c *Client) GetVirtualHostNode(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("virtualhostnode/" + url.PathEscape(name))
}

// DeleteVirtualHostNode ...
func (c *Client) DeleteVirtualHostNode(name string) (res *http.Response, err error) {
	return c.deleteConfiguredObject("virtualhostnode/" + url.PathEscape(name))
}

// UpdateVirtualHostNode ...
func (c *Client) UpdateVirtualHostNode(name string, attributes *map[string]interface{}) (res *http.Response, err error) {
	return c.restClient.Post("virtualhostnode/"+url.PathEscape(name), attributes)
}

// GetVirtualHost ...
func (c *Client) GetVirtualHost(node string, host string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("virtualhost/" + url.PathEscape(node) + "/" + url.PathEscape(host))
}

// CreateVirtualHost ...
func (c *Client) CreateVirtualHost(node string, attributes *map[string]interface{}) (res *http.Response, err error) {
	return c.restClient.Post("virtualhost/"+url.PathEscape(node), attributes)
}

// DeleteVirtualHost ...
func (c *Client) DeleteVirtualHost(node string, host string) (res *http.Response, err error) {
	return c.deleteConfiguredObject("virtualhost/" + url.PathEscape(node) + "/" + url.PathEscape(host))
}

// UpdateVirtualHost
func (c *Client) UpdateVirtualHost(node string, name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("virtualhost/"+url.PathEscape(node)+"/"+url.PathEscape(name), attributes)
}

func (c *Client) getConfiguredObject(path string) (*map[string]interface{}, error) {
	return c.getConfiguredObjectAttributes(path, true)
}

func (c *Client) getConfiguredObjectAttributes(path string, actuals bool) (*map[string]interface{}, error) {
	v := url.Values{}
	v.Set("actuals", strconv.FormatBool(actuals))
	return c.restClient.GetAsMap(path, v)
}

func (c *Client) deleteConfiguredObject(path string) (*http.Response, error) {
	return c.restClient.Delete(path)
}

// CreateQueue ...
func (c *Client) CreateQueue(node string, host string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("queue/"+url.PathEscape(node)+"/"+url.PathEscape(host), attributes)
}

// GetQueue ...
func (c *Client) GetQueue(node string, host string, name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("queue/" + url.PathEscape(node) + "/" + url.PathEscape(host) + "/" + url.PathEscape(name))
}

// DeleteQueue ...
func (c *Client) DeleteQueue(node string, host string, name string) (res *http.Response, err error) {
	return c.deleteConfiguredObject("queue/" + url.PathEscape(node) + "/" + url.PathEscape(host) + "/" + url.PathEscape(name))
}

// UpdateQueue ...
func (c *Client) UpdateQueue(node string, host string, name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("queue/"+url.PathEscape(node)+"/"+url.PathEscape(host)+"/"+url.PathEscape(name), attributes)
}

// CreateExchange...
func (c *Client) CreateExchange(node string, host string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("exchange/"+url.PathEscape(node)+"/"+url.PathEscape(host), attributes)
}

// GetExchange ...
func (c *Client) GetExchange(node string, host string, name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("exchange/" + url.PathEscape(node) + "/" + url.PathEscape(host) + "/" + url.PathEscape(name))
}

// DeleteExchange ...
func (c *Client) DeleteExchange(node string, host string, name string) (res *http.Response, err error) {
	return c.deleteConfiguredObject("exchange/" + url.PathEscape(node) + "/" + url.PathEscape(host) + "/" + url.PathEscape(name))
}

// UpdateExchange ...
func (c *Client) UpdateExchange(node string, host string, name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("exchange/"+url.PathEscape(node)+"/"+url.PathEscape(host)+"/"+url.PathEscape(name), attributes)
}

func (c *Client) CreateBinding(b *Binding) (*http.Response, error) {

	return c.makeBinding(b, false)
}


func (c *Client) UpdateBinding(b *Binding) (*http.Response, error) {

	return c.makeBinding(b, true)
}

func (c *Client) GetBinding(b *Binding) (*Binding, error) {
	path := "exchange/" + url.PathEscape(b.VirtualHostNode) + "/" + url.PathEscape(b.VirtualHost) + "/" + url.PathEscape(b.Exchange )
	attributes, err := c.getConfiguredObjectAttributes(path, false)
	if err != nil {
		return &Binding{}, err
	}

	if len(*attributes) == 0 {
		return nil, nil
	}
	bindings := (*attributes)["bindings"]

	log.Printf("exchange %s bindings %v ",b.Exchange , bindings )
	if bindings != nil {
		bs := bindings.([]interface{})
		for _, bn := range bs {
			bnd := bn.(map[string]interface{})
			log.Printf("binding %v for destination %v",bnd["name"], bnd["destination"] )
			if bnd["name"] == b.BindingKey && bnd["destination"] == b.Destination {
				args := bnd["arguments"]
				var arguments map[string]string
				if args != nil {
					i := args.(map[string]interface{})
					arguments = *convertToMapOfStrings(&i)
				}
				return &Binding{
					b.BindingKey,
					b.Destination,
					b.Exchange,
					arguments,
					b.VirtualHostNode,
					b.VirtualHost}, nil
			}
		}
	}
	return nil, nil
}

func (c *Client) makeBinding(b *Binding, replaceExistingArguments bool) (*http.Response, error) {
	var arguments = &map[string]interface{}{
		"destination":              b.Destination,
		"bindingKey":               b.BindingKey,
		"arguments":                b.Arguments,
		"replaceExistingArguments": replaceExistingArguments}
	return c.restClient.Post("exchange/"+url.PathEscape(b.VirtualHostNode)+"/"+url.PathEscape(b.VirtualHost)+"/"+url.PathEscape(b.Exchange) + "/bind", arguments)
}


func (c *Client) DeleteBinding(b *Binding) (*http.Response, error) {
	var arguments = &map[string]interface{}{
		"destination":              b.Destination,
		"bindingKey":               b.BindingKey}
	return c.restClient.Post("exchange/"+url.PathEscape(b.VirtualHostNode)+"/"+url.PathEscape(b.VirtualHost)+"/"+url.PathEscape(b.Exchange) + "/unbind", arguments)
}