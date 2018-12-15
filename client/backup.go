package main

import (
	"io"
)

// Backup backups all blog database.
func (c *Client) Backup(w io.Writer) {
	resp := c.mustGet("/backups")
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}
