package client

import (
	"github.com/movsb/taoblog/protocols/clients"
)

type Client struct {
	*clients.ProtoClient
}

// NewClient creates a new client that interacts with server.
func NewClient(config HostConfig) *Client {
	c := &Client{
		ProtoClient: clients.NewFromHome(config.Home, config.Token),
	}
	return c
}
