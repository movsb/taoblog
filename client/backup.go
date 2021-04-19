package client

import (
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/movsb/taoblog/protocols"
	"github.com/spf13/cobra"
)

// Backup backups all blog database.
func (c *Client) Backup(cmd *cobra.Command) {
	compress := true

	backupClient, err := c.management.Backup(c.token(), &protocols.BackupRequest{Compress: compress})
	if err != nil {
		panic(err)
	}

	bpr := &_BackupProgressReader{c: backupClient}
	defer backupClient.CloseSend()

	var r io.ReadCloser
	if compress {
		zr, err := zlib.NewReader(bpr)
		if err != nil {
			panic(err)
		}
		r = zr
	} else {
		r = ioutil.NopCloser(bpr)
	}
	defer r.Close()

	var w io.Writer
	bStdout, err := cmd.Flags().GetBool(`stdout`)
	if err != nil {
		panic(err)
	}
	if !bStdout {
		name := time.Now().Format(`taoblog-2006-01-02.db`)
		fp, err := os.Create(name)
		if err != nil {
			panic(err)
		}
		defer fmt.Printf("Filename: %s\n", name)
		defer fp.Close()
		w = fp
	} else {
		w = os.Stdout
	}

	if _, err := io.Copy(w, r); err != nil {
		panic(err)
	}

	fmt.Println()
}

type _BackupProgressReader struct {
	c         protocols.Management_BackupClient
	d         []byte
	preparing bool
}

// Read ...
func (r *_BackupProgressReader) Read(p []byte) (int, error) {
	if len(r.d) == 0 {
		rsp, err := r.c.Recv()
		if err != nil {
			if err == io.EOF {
				return 0, err
			}
			panic(err)
		}
		switch typed := rsp.BackupResponseMessage.(type) {
		case *protocols.BackupResponse_Preparing_:
			fmt.Printf("\r\033[KPreparing... %.2f%%", typed.Preparing.Progress*100)
			r.preparing = true
		case *protocols.BackupResponse_Transfering_:
			if r.preparing {
				fmt.Println()
				r.preparing = false
			}
			fmt.Printf("\r\033[KTransfering... %.2f%%", typed.Transfering.Progress*100)
			r.d = typed.Transfering.Data
		}
	}

	n := copy(p, r.d)
	r.d = r.d[n:]
	return n, nil
}
