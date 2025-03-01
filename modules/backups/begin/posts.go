package begin

import (
	"io"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type BackupClient struct {
	cc *clients.ProtoClient
}

func NewBackupClient(cc *clients.ProtoClient) *BackupClient {
	return &BackupClient{
		cc: cc,
	}
}

func (b *BackupClient) BackupPosts(w io.Writer) (outErr error) {
	defer utils.CatchAsError(&outErr)

	client := utils.Must1(b.cc.Management.Backup(
		b.cc.Context(),
		&proto.BackupRequest{Compress: false},
	))

	bpr := &_BackupProgressReader{c: client}
	return utils.KeepLast1(io.Copy(w, bpr))
}

type _BackupProgressReader struct {
	c proto.Management_BackupClient
	d []byte
}

func (r *_BackupProgressReader) Read(p []byte) (outN int, outErr error) {
	defer utils.CatchAsError(&outErr)

	if len(r.d) == 0 {
		rsp := utils.Must1(r.c.Recv())
		switch typed := rsp.BackupResponseMessage.(type) {
		case *proto.BackupResponse_Transfering_:
			r.d = typed.Transfering.Data
		}
	}

	n := copy(p, r.d)
	r.d = r.d[n:]
	return n, nil
}
