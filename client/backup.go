package main

import (
	"bytes"
	"context"
	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc/metadata"
	"io"
)

// Backup backups all blog database.
func (c *Client) Backup(w io.Writer) {
	ctx := metadata.NewOutgoingContext(context.TODO(), metadata.Pairs("token", c.config.Token))
	resp, err := c.grpcClient.Backup(ctx, &protocols.BackupRequest{})
	if err != nil {
		panic(err)
	}
	if _, err := io.Copy(w, bytes.NewBufferString(resp.Data)); err != nil {
		panic(err)
	}
}
