package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Client ...
// TODO: close client connection.
type Client struct {
	config HostConfig

	client *http.Client
	cc     *grpc.ClientConn

	blog       protocols.TaoBlogClient
	management protocols.ManagementClient
}

// NewClient creates a new client that interacts with server.
func NewClient(config HostConfig) *Client {
	c := &Client{
		config: config,
	}
	c.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !config.Verify,
			},
		},
	}

	grpcAddress := c.config.GRPC
	secure := false
	if grpcAddress == `` {
		u, _ := url.Parse(c.config.API)
		grpcAddress = u.Hostname()
		grpcPort := u.Port()
		if u.Scheme == `http` {
			secure = false
			if grpcPort == `` {
				grpcPort = `80`
			}
		} else {
			secure = true
			if grpcPort == `` {
				grpcPort = `443`
			}
		}

		grpcAddress = fmt.Sprintf(`%s:%s`, grpcAddress, grpcPort)
	}

	var conn *grpc.ClientConn
	var err error
	if secure {
		if conn, err = grpc.Dial(
			grpcAddress,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100<<20)),
		); err != nil {
			panic(err)
		}
	} else {
		if conn, err = grpc.Dial(
			grpcAddress,
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100<<20)),
		); err != nil {
			panic(err)
		}
	}

	c.cc = conn

	c.blog = protocols.NewTaoBlogClient(c.cc)
	c.management = protocols.NewManagementClient(c.cc)

	return c
}

func (c *Client) token() context.Context {
	return metadata.NewOutgoingContext(context.TODO(), metadata.Pairs(auth.TokenName, c.config.Token))
}
