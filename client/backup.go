package client

import (
	"bytes"
	"compress/zlib"
	"io"
	"io/ioutil"

	"github.com/movsb/taoblog/protocols"
)

// Backup backups all blog database.
func (c *Client) Backup(w io.Writer) {
	compress := true

	resp, err := c.grpcClient.Backup(c.token(), &protocols.BackupRequest{Compress: compress})
	if err != nil {
		panic(err)
	}

	var reader io.ReadCloser

	if compress {
		r, err := zlib.NewReader(bytes.NewReader(resp.Data))
		if err != nil {
			panic(err)
		}
		reader = r
	} else {
		reader = ioutil.NopCloser(bytes.NewReader(resp.Data))
	}

	defer reader.Close()

	if _, err := io.Copy(w, reader); err != nil {
		panic(err)
	}
}
