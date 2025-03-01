package backups

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"

	"github.com/movsb/taoblog/modules/backups/begin"
	backups_crypto "github.com/movsb/taoblog/modules/backups/crypto"
	"github.com/movsb/taoblog/modules/backups/r2"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
)

type Remote interface {
	Upload(ctx context.Context, path string, r io.Reader) error
}

type Backup struct {
	ctx   context.Context
	cc    *clients.ProtoClient
	store utils.PluginStorage

	remote   Remote
	identity string
}

type Option func(b *Backup)

func WithRemoteR2(accountID, accessKeyID, accessKeySecret, bucketName string) Option {
	return func(b *Backup) {
		b.remote = utils.Must1(r2.New(accountID, accessKeyID, accessKeySecret, bucketName))
	}
}

func WithEncoderAge(identity string) Option {
	return func(b *Backup) {
		b.identity = identity
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

	if b.identity == `` {
		panic(`没有指定私钥。`)
	}

	return &b, nil
}

type _RW struct {
	b *bytes.Buffer
	w io.Writer

	closers []io.Closer
}

func (r *_RW) Writer() io.Writer {
	return r.w
}
func (r *_RW) Close() (_ io.ReadSeeker, outErr error) {
	defer utils.CatchAsError(&outErr)
	for _, c := range r.closers {
		utils.Must(c.Close())
	}
	return bytes.NewReader(r.b.Bytes()), nil
}

func (b *Backup) createWriter() (_ *_RW, outErr error) {
	defer utils.CatchAsError(&outErr)

	buf := bytes.NewBuffer(nil)
	aw := utils.Must1(backups_crypto.NewAge(b.identity, buf))
	// 总量压缩
	gw := gzip.NewWriter(aw)

	closers := []io.Closer{gw, aw}

	return &_RW{b: buf, w: gw, closers: closers}, nil
}

// TODO 使用临时文件缓存替代内存
func (b *Backup) BackupPosts(ctx context.Context) (outErr error) {
	defer utils.CatchAsError(&outErr)
	bb := begin.NewBackupClient(b.cc)
	wc := utils.Must1(b.createWriter())
	utils.Must(bb.BackupPosts(wc.Writer()))
	r := utils.Must1(wc.Close())
	return b.remote.Upload(ctx, `posts.db.gz.age`, r)
}
