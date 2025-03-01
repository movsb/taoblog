package backups

import (
	"bytes"
	"context"
	"io"

	"github.com/movsb/taoblog/modules/backups/begin"
	"github.com/movsb/taoblog/modules/backups/r2"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
)

type Remote interface {
	Upload(ctx context.Context, path string, r io.Reader) error
}

type Backup struct {
	ctx    context.Context
	cc     *clients.ProtoClient
	remote Remote
	store  utils.PluginStorage
}

type Option func(b *Backup)

func WithRemoteR2(accountID, accessKeyID, accessKeySecret, bucketName string) Option {
	return func(b *Backup) {
		b.remote = utils.Must1(r2.New(accountID, accessKeyID, accessKeySecret, bucketName))
	}
}

// NOTE: grpc stream 无法直接使用 Server，只能从地址注册 client 使用
// NOTE：然而 storage 又是本地进程的。
func New(ctx context.Context, store utils.PluginStorage, grpcAddress string, options ...Option) (outB *Backup, outErr error) {
	defer utils.CatchAsError(&outErr)

	b := Backup{
		ctx:   ctx,
		store: store,
		cc:    clients.NewProtoClientAsSystemAdmin(grpcAddress),
	}

	for _, opt := range options {
		opt(&b)
	}

	if b.remote == nil {
		panic(`没有指定存储后端。`)
	}

	return &b, nil
}

func (b *Backup) BackupPosts(ctx context.Context) (outErr error) {
	defer utils.CatchAsError(&outErr)
	bb := begin.NewBackupClient(b.cc)
	buf := bytes.NewBuffer(nil)
	utils.Must(bb.BackupPosts(buf))
	seekBuf := bytes.NewReader(buf.Bytes())
	return b.remote.Upload(ctx, `posts.db`, seekBuf)
}
