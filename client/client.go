package main

import (
	"crypto/tls"
	"errors"
	"net/http"
)

var (
	// ErrStatusCode ...
	ErrStatusCode = errors.New("http.statusCode != 200")
)

// Client ...
type Client struct {
	client *http.Client
}

// NewClient ...
func NewClient() *Client {
	c := &Client{}
	c.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !initConfig.verify,
			},
		},
	}
	return c
}

func (c *Client) get(path string) *http.Response {
	req, err := http.NewRequest("GET", initConfig.api+path, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", initConfig.key)
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
