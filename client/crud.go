package main

import (
	"fmt"
	"io"
	"os"
)

// CRUD ...
func (c *Client) CRUD(method string, uri string) {
	if method == "get" {
		resp := c.get(uri)
		fmt.Println(resp.Proto, resp.Status)
		resp.Header.Write(os.Stdout)
		fmt.Println()
		defer resp.Body.Close()
		io.Copy(os.Stdout, resp.Body)
		fmt.Println()
	}
}
