package client

import (
	"bytes"
	"io"

	"github.com/movsb/taoblog/protocols"
)

// Backup backups all blog database.
func (c *Client) Backup(w io.Writer) {
	resp, err := c.grpcClient.Backup(c.token(), &protocols.BackupRequest{})
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(w, bytes.NewBufferString(resp.Data)); err != nil {
		panic(err)
	}
}
