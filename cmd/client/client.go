package client

import (
	"github.com/movsb/taoblog/protocols"
)

// Client ...
// TODO: close client connection.
type Client struct {
	config HostConfig
	*proto.ProtoClient
}

// NewClient creates a new client that interacts with server.
func NewClient(config HostConfig) *Client {
	c := &Client{
		config: config,
		ProtoClient: proto.NewProtoClient(
			proto.NewConn(config.API, config.GRPC),
			config.Token,
		),
	}

	return c
}
