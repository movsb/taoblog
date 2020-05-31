package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
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
	contentTypeJSON   = "application/json"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// Client ...
type Client struct {
	config     HostConfig
	client     *http.Client
	grpcClient protocols.TaoBlogClient
}

// NewClient ...
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
		); err != nil {
			panic(err)
		}
	} else {
		if conn, err = grpc.Dial(
			grpcAddress,
			grpc.WithInsecure(),
		); err != nil {
			panic(err)
		}
	}

	c.grpcClient = protocols.NewTaoBlogClient(conn)
	return c
}

func (c *Client) token() context.Context {
	return metadata.NewOutgoingContext(context.TODO(), metadata.Pairs("token", c.config.Token))
}

func (c *Client) get(path string) *http.Response {
	req, err := http.NewRequest("GET", c.config.API+path, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", c.config.Token)
	resp, err := c.client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func (c *Client) mustGet(path string) *http.Response {
	resp := c.get(path)
	if resp.StatusCode != 200 {
		resp.Body.Close()
		panic(ErrStatusCode)
	}
	return resp
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

func (c *Client) mustPostJSON(path string, data interface{}) *http.Response {
	bys, _ := json.Marshal(data)
	resp := c.post(path, bytes.NewReader(bys), contentTypeJSON)
	if resp.StatusCode != 200 {
		io.Copy(os.Stderr, resp.Body)
		resp.Body.Close()
		panic(resp.Status)
	}
	return resp
}
