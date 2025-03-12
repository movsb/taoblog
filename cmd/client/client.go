package client

import (
	"github.com/movsb/taoblog/protocols/clients"
)

// Client ...
// TODO: close client connection.
type Client struct {
	config HostConfig
	*clients.ProtoClient
}

// NewClient creates a new client that interacts with server.
func NewClient(config HostConfig) *Client {
	c := &Client{
		config:      config,
		ProtoClient: clients.NewProtoClientFromHome(config.Home, config.Token),
	}

	return c
}
