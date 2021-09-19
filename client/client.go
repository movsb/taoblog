package client

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	// ErrStatusCode ...
	ErrStatusCode = errors.New("http.statusCode != 200")
)

const (
	contentTypeBinary = "application/octet-stream"
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
	return metadata.NewOutgoingContext(context.TODO(), metadata.Pairs("token", c.config.Token))
}

func (c *Client) post(path string, body io.Reader, ty string) *http.Response {
	req, err := http.NewRequest("POST", c.config.API+path, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", c.config.Token)
	req.Header.Set("Content-Type", ty)
	resp, err := c.client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func (c *Client) mustPost(path string, body io.Reader, ty string) *http.Response {
	resp := c.post(path, body, ty)
	if resp.StatusCode != 200 {
		io.Copy(os.Stderr, resp.Body)
		resp.Body.Close()
		panic(resp.Status)
	}
	return resp
}
