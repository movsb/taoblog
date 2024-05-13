package client

import (
	"context"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Client ...
// TODO: close client connection.
type Client struct {
	config HostConfig

	cc *grpc.ClientConn

	blog       protocols.TaoBlogClient
	management protocols.ManagementClient
}

// NewClient creates a new client that interacts with server.
func NewClient(config HostConfig) *Client {
	c := &Client{
		config: config,
	}

	c.cc = NewConn(c.config.API, c.config.GRPC)
	c.blog = protocols.NewTaoBlogClient(c.cc)
	c.management = protocols.NewManagementClient(c.cc)

	return c
}

func (c *Client) token() context.Context {
	return metadata.NewOutgoingContext(context.TODO(), metadata.Pairs(auth.TokenName, c.config.Token))
}

func (c *Client) Token() context.Context {
	return c.token()
}

func (c *Client) Blog() protocols.TaoBlogClient {
	return c.blog
}
