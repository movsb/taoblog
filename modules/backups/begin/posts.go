package begin

import (
	"io"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/setup/migration"
)

type BackupClient struct {
	cc *clients.ProtoClient
}

func NewBackupClient(cc *clients.ProtoClient) *BackupClient {
	return &BackupClient{
		cc: cc,
	}
}

func (b *BackupClient) BackupDatabase(w io.Writer) (outErr error) {
	defer utils.CatchAsError(&outErr)

	client := utils.Must1(b.cc.Management.Backup(
		b.cc.Context(),
		&proto.BackupRequest{
			ClientDatabaseVersion: int32(migration.MaxVersionNumber()),
			Compress:              false,
		},
	))
	defer client.CloseSend()

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
		case *proto.BackupResponse_Transferring_:
			r.d = typed.Transferring.Data
		}
	}

	n := copy(p, r.d)
	r.d = r.d[n:]
	return n, nil
}

func (b *BackupClient) BackupFiles(postID int, writeFile func(spec *proto.FileSpec, r io.Reader) error) (outErr error) {
	defer utils.CatchAsError(&outErr)

	fileList := utils.Must1(b.cc.Blog.ListPostFiles(
		b.cc.Context(),
		&proto.ListPostFilesRequest{
			PostId:             int32(postID),
			WithGenerated:      true,
			WithLivePhotoVideo: true,
		},
	)).GetFiles()

	for _, spec := range fileList {
		func(spec *proto.FileSpec) {
			client := utils.Must1(b.cc.Blog.GetPostFileStream(b.cc.Context(), &proto.GetPostFileStreamRequest{
				PostId: int32(postID),
				Path:   spec.Path,
			}))
			defer client.CloseSend()
			utils.Must(writeFile(spec, &_FileReader{c: client}))
		}(spec)
	}

	return nil
}

type _FileReader struct {
	c   proto.TaoBlog_GetPostFileStreamClient
	d   []byte
	eof bool
}

func (r *_FileReader) Read(p []byte) (outN int, outErr error) {
	defer utils.CatchAsError(&outErr)

	if len(r.d) == 0 {
		if r.eof {
			return 0, io.EOF
		}
		rsp := utils.Must1(r.c.Recv())
		// 可以保留指针吗？
		r.d = rsp.GetData()
		r.eof = len(r.d) == 0
	}

	if r.eof {
		return 0, io.EOF
	}

	n := copy(p, r.d)
	r.d = r.d[n:]
	return n, nil
}
