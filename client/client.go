package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/movsb/taoblog/modules/stdinlinereader"
	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc"
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
	line       *stdinlinereader.StdinLineReader
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
	c.line = stdinlinereader.NewStdinLineReader()

	u, _ := url.Parse(c.config.API)

	conn, err := grpc.Dial(u.Hostname()+":2563", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	c.grpcClient = protocols.NewTaoBlogClient(conn)
	return c
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
		panic(ErrStatusCode)
	}
	return resp
}

func (c *Client) mustPostJSON(path string, data interface{}) *http.Response {
	bys, _ := json.Marshal(data)
	resp := c.post(path, bytes.NewReader(bys), contentTypeJSON)
	if resp.StatusCode != 200 {
		resp.Body.Close()
		panic(ErrStatusCode)
	}
	return resp
}
