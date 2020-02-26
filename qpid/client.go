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
	bindings, err := c.getExchangeBindings(b.VirtualHostNode, b.VirtualHost, b.Exchange)
	if err != nil {
		return &Binding{}, err
	}

	if len(*bindings) == 0 {
		return nil, nil
	}

	for _, bnd := range *bindings {
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

	return nil, nil
}

func (c *Client) makeBinding(b *Binding, replaceExistingArguments bool) (*http.Response, error) {
	var arguments = &map[string]interface{}{
		"destination":              b.Destination,
		"bindingKey":               b.BindingKey,
		"arguments":                b.Arguments,
		"replaceExistingArguments": replaceExistingArguments}
	return c.restClient.Post("exchange/"+url.PathEscape(b.VirtualHostNode)+"/"+url.PathEscape(b.VirtualHost)+"/"+url.PathEscape(b.Exchange)+"/bind", arguments)
}

func (c *Client) DeleteBinding(b *Binding) (*http.Response, error) {
	var arguments = &map[string]interface{}{
		"destination": b.Destination,
		"bindingKey":  b.BindingKey}
	return c.restClient.Post("exchange/"+url.PathEscape(b.VirtualHostNode)+"/"+url.PathEscape(b.VirtualHost)+"/"+url.PathEscape(b.Exchange)+"/unbind", arguments)
}

func (c *Client) listConfiguredObjets(path string, actuals bool) (*[]map[string]interface{}, error) {
	v := url.Values{}
	v.Set("actuals", strconv.FormatBool(actuals))
	return c.restClient.GetAsArray(path, v)
}

func (c *Client) GetVirtualHostNodes() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("virtualhostnode", true)
}

func (c *Client) GetVirtualHosts() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("virtualhost", true)
}

func (c *Client) GetNodeVirtualHosts(nodeName string) (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("virtualhost/"+url.PathEscape(nodeName), true)
}

func (c *Client) getQueues() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("queue", true)
}

func (c *Client) getVirtualHostQueues(nodeName string, hostName string) (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("queue/"+url.PathEscape(nodeName)+"/"+url.PathEscape(hostName), true)
}

func (c *Client) getVirtualHostExchanges(nodeName string, hostName string) (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("exchange/"+url.PathEscape(nodeName)+"/"+url.PathEscape(hostName), true)
}

func (c *Client) getExchangeBindings(nodeName string, hostName string, exchange string) (*[]map[string]interface{}, error) {
	path := "exchange/" + url.PathEscape(nodeName) + "/" + url.PathEscape(hostName) + "/" + url.PathEscape(exchange)
	attributes, err := c.getConfiguredObjectAttributes(path, false)
	if err != nil {
		return &[]map[string]interface{}{}, err
	}

	if len(*attributes) == 0 {
		return &[]map[string]interface{}{}, nil
	}

	bindings := (*attributes)["bindings"]

	if bindings != nil {
		bs := bindings.([]interface{})
		result := make([]map[string]interface{}, len(bs))
		for idx, bn := range bs {
			bnd := bn.(map[string]interface{})
			result[idx] = bnd
		}
		return &result, nil
	}
	return &[]map[string]interface{}{}, nil
}

func (c *Client) CreateAuthenticationProvider(attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("authenticationprovider", attributes)
}

func (c *Client) GetAuthenticationProvider(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("authenticationprovider/" + url.PathEscape(name))
}

func (c *Client) DeleteAuthenticationProvider(name string) (*http.Response, error) {
	return c.deleteConfiguredObject("authenticationprovider/" + url.PathEscape(name))
}

func (c *Client) UpdateAuthenticationProvider(name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("authenticationprovider/"+url.PathEscape(name), attributes)
}

func (c *Client) GetAuthenticationProviders() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("authenticationprovider", true)
}

func (c *Client) CreateUser(authenticationProvider string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("user/"+url.PathEscape(authenticationProvider), attributes)
}

func (c *Client) GetUser(authenticationProvider string, name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("user/" + url.PathEscape(authenticationProvider) + "/" + url.PathEscape(name))
}

func (c *Client) DeleteUser(authenticationProvider string, name string) (*http.Response, error) {
	return c.deleteConfiguredObject("user/" + url.PathEscape(authenticationProvider) + "/" + url.PathEscape(name))
}

func (c *Client) UpdateUser(authenticationProvider string, name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("user/"+url.PathEscape(authenticationProvider)+"/"+url.PathEscape(name), attributes)
}

func (c *Client) GetUsers(authenticationProvider string) (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("user/"+url.PathEscape(authenticationProvider), true)
}

func (c *Client) CreateGroupProvider(attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("groupprovider", attributes)
}

func (c *Client) GetGroupProvider(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("groupprovider/" + url.PathEscape(name))
}

func (c *Client) DeleteGroupProvider(name string) (*http.Response, error) {
	return c.deleteConfiguredObject("groupprovider/" + url.PathEscape(name))
}

func (c *Client) UpdateGroupProvider(name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("groupprovider/"+url.PathEscape(name), attributes)
}

func (c *Client) GetGroupProviders() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("groupprovider", true)
}

func (c *Client) CreateGroup(groupProvider string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("group/"+url.PathEscape(groupProvider), attributes)
}

func (c *Client) GetGroup(groupProvider string, name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("group/" + url.PathEscape(groupProvider) + "/" + url.PathEscape(name))
}

func (c *Client) DeleteGroup(groupProvider string, name string) (*http.Response, error) {
	return c.deleteConfiguredObject("group/" + url.PathEscape(groupProvider) + "/" + url.PathEscape(name))
}

func (c *Client) UpdateGroup(groupProvider string, name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("group/"+url.PathEscape(groupProvider)+"/"+url.PathEscape(name), attributes)
}

func (c *Client) GetGroups(groupProvider string) (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("group/"+url.PathEscape(groupProvider), true)
}

func (c *Client) CreateGroupMember(groupProvider string, groupName string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("groupmember/"+url.PathEscape(groupProvider)+"/"+url.PathEscape(groupName), attributes)
}

func (c *Client) GetGroupMember(groupProvider string, groupName string, name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("groupmember/" + url.PathEscape(groupProvider) + "/" + url.PathEscape(groupName) + "/" + url.PathEscape(name))
}

func (c *Client) DeleteGroupMember(groupProvider string, groupName string, name string) (*http.Response, error) {
	return c.deleteConfiguredObject("groupmember/" + url.PathEscape(groupProvider) + "/" + url.PathEscape(groupName) + "/" + url.PathEscape(name))
}

func (c *Client) UpdateGroupMember(groupProvider string, groupName string, name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("groupmember/"+url.PathEscape(groupProvider)+"/"+url.PathEscape(groupName)+"/"+url.PathEscape(name), attributes)
}

func (c *Client) GetGroupMembers(groupProvider string, groupName string) (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("groupmember/"+url.PathEscape(groupProvider)+"/"+url.PathEscape(groupName), true)
}

func (c *Client) CreateAccessControlProvider(attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("accesscontrolprovider", attributes)
}

func (c *Client) GetAccessControlProvider(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("accesscontrolprovider/" + url.PathEscape(name))
}

func (c *Client) DeleteAccessControlProvider(name string) (*http.Response, error) {
	return c.deleteConfiguredObject("accesscontrolprovider/" + url.PathEscape(name))
}

func (c *Client) UpdateAccessControlProvider(name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("accesscontrolprovider/"+url.PathEscape(name), attributes)
}

func (c *Client) GetAccessControlProviders() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("accesscontrolprovider", true)
}

func (c *Client) CreateKeyStore(attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("keystore", attributes)
}

func (c *Client) GetKeyStore(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("keystore/" + url.PathEscape(name))
}

func (c *Client) DeleteKeyStore(name string) (*http.Response, error) {
	return c.deleteConfiguredObject("keystore/" + url.PathEscape(name))
}

func (c *Client) UpdateKeyStore(name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("keystore/"+url.PathEscape(name), attributes)
}

func (c *Client) GetKeyStores() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("keystore", true)
}

func (c *Client) CreateTrustStore(attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("truststore", attributes)
}

func (c *Client) GetTrustStore(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("truststore/" + url.PathEscape(name))
}

func (c *Client) DeleteTrustStore(name string) (*http.Response, error) {
	return c.deleteConfiguredObject("truststore/" + url.PathEscape(name))
}

func (c *Client) UpdateTrustStore(name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("truststore/"+url.PathEscape(name), attributes)
}

func (c *Client) GetTrustStores() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("truststore", true)
}

func (c *Client) CreatePort(attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("port", attributes)
}

func (c *Client) GetPort(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("port/" + url.PathEscape(name))
}

func (c *Client) DeletePort(name string) (*http.Response, error) {
	return c.deleteConfiguredObject("port/" + url.PathEscape(name))
}

func (c *Client) UpdatePort(name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("port/"+url.PathEscape(name), attributes)
}

func (c *Client) GetPorts() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("port", true)
}

func (c *Client) CreateVirtualHostAlias(portName string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("virtualhostalias/"+url.PathEscape(portName), attributes)
}

func (c *Client) GetVirtualHostAlias(portName string, name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("virtualhostalias/" + url.PathEscape(portName) + "/" + url.PathEscape(name))
}

func (c *Client) DeleteVirtualHostAlias(portName string, name string) (*http.Response, error) {
	return c.deleteConfiguredObject("virtualhostalias/" + url.PathEscape(portName) + "/" + url.PathEscape(name))
}

func (c *Client) UpdateVirtualHostAlias(portName string, name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("virtualhostalias/"+url.PathEscape(portName)+"/"+url.PathEscape(name), attributes)
}

func (c *Client) GetVirtualHostAliases(portName string) (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("virtualhostalias/"+url.PathEscape(portName), true)
}

func (c *Client) CreateBrokerLogger(attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("brokerlogger", attributes)
}

func (c *Client) GetBrokerLogger(name string) (*map[string]interface{}, error) {
	return c.getConfiguredObject("brokerlogger/" + url.PathEscape(name))
}

func (c *Client) DeleteBrokerLogger(name string) (*http.Response, error) {
	return c.deleteConfiguredObject("brokerlogger/" + url.PathEscape(name))
}

func (c *Client) UpdateBrokerLogger(name string, attributes *map[string]interface{}) (*http.Response, error) {
	return c.restClient.Post("brokerlogger/"+url.PathEscape(name), attributes)
}

func (c *Client) GetBrokerLoggers() (*[]map[string]interface{}, error) {
	return c.listConfiguredObjets("brokerlogger", true)
}
