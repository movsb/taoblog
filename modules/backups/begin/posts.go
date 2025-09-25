package begin

import (
	"bytes"
	"io"
	"log"

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

func (b *BackupClient) Backup(w io.Writer) (outErr error) {
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

	client := utils.Must1(b.cc.Management.FileSystem(b.cc.Context()))
	defer client.CloseSend()

	utils.Must(client.Send(&proto.FileSystemRequest{
		Init: &proto.FileSystemRequest_InitRequest{
			For: &proto.FileSystemRequest_InitRequest_Post_{
				Post: &proto.FileSystemRequest_InitRequest_Post{
					Id: int64(postID),
				},
			},
		},
	}))
	if utils.Must1(client.Recv()).GetInit() == nil {
		log.Panicln(`文件系统初始化失败。`)
	}

	utils.Must(client.Send(&proto.FileSystemRequest{
		Request: &proto.FileSystemRequest_ListFiles{
			ListFiles: &proto.FileSystemRequest_ListFilesRequest{},
		},
	}))

	files := utils.Must1(client.Recv()).GetListFiles().GetFiles()
	for _, file := range files {
		utils.Must(writeFile(file, utils.Must1(_NewFileRead(client, file.Path))))
	}
	return nil
}

type _FileReader struct {
	io.Reader
}

func _NewFileRead(client proto.Management_FileSystemClient, path string) (_ *_FileReader, outErr error) {
	defer utils.CatchAsError(&outErr)
	utils.Must(client.Send(&proto.FileSystemRequest{
		Request: &proto.FileSystemRequest_ReadFile{
			ReadFile: &proto.FileSystemRequest_ReadFileRequest{
				Path: path,
			},
		},
	}))
	data := utils.Must1(client.Recv()).GetReadFile().Data
	return &_FileReader{Reader: bytes.NewReader(data)}, nil
}
