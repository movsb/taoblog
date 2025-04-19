package backups

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/movsb/taoblog/modules/backups/begin"
	backups_crypto "github.com/movsb/taoblog/modules/backups/crypto"
	"github.com/movsb/taoblog/modules/backups/r2"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type Remote interface {
	Upload(ctx context.Context, path string, r io.Reader, contentType string) error
}

type Backup struct {
	ctx    context.Context
	client *clients.ProtoClient
	store  utils.PluginStorage

	remote   Remote
	identity string

	// 由于 R2 目前好像没有版本机制，所以旧文件会覆盖新文件，
	// 为了避免如此，每次上传为不同的文件名，循环覆盖。
	// 如果要手动恢复，取修改时间最新的即可。
	nextPostsFileIndex int
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
func New(ctx context.Context, store utils.PluginStorage, client *clients.ProtoClient, options ...Option) (outB *Backup, outErr error) {
	defer utils.CatchAsError(&outErr)

	b := Backup{
		ctx:    ctx,
		store:  store,
		client: client,
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
	// 总是压缩
	gw := gzip.NewWriter(aw)

	closers := []io.Closer{gw, aw}

	return &_RW{b: buf, w: gw, closers: closers}, nil
}

// TODO 使用临时文件缓存替代内存
func (b *Backup) BackupPosts(ctx context.Context) (outErr error) {
	defer utils.CatchAsError(&outErr)
	bb := begin.NewBackupClient(b.client)
	wc := utils.Must1(b.createWriter())
	utils.Must(bb.BackupPosts(wc.Writer()))
	r := utils.Must1(wc.Close())

	// 1, 2, 3... 循环命名。
	next := b.nextPostsFileIndex%3 + 1
	b.nextPostsFileIndex++
	name := fmt.Sprintf(`posts-%d.db.gz.age`, next)

	return b.remote.Upload(ctx, name, r, ``)
}

func (b *Backup) BackupFiles(ctx context.Context) (outErr error) {
	defer utils.CatchAsError(&outErr)

	lastTime, err := b.store.GetIntegerDefault(`last_time`, 0)
	if err != nil {
		log.Panicln(`获取上次更新时间失败：`, err)
	}

	now := time.Now()

	updatedPosts := utils.Must1(b.client.Blog.ListPosts(
		b.client.Context(),
		&proto.ListPostsRequest{
			ModifiedNotBefore: int32(lastTime),
		},
	)).GetPosts()

	bb := begin.NewBackupClient(b.client)
	for _, post := range updatedPosts {
		select {
		case <-ctx.Done():
			log.Println(`终止备份`)
			return
		default:
		}
		wc := utils.Must1(b.createWriter())
		tw := tar.NewWriter(wc.Writer())

		fileCount := 0

		utils.Must(bb.BackupFiles(int(post.Id), func(spec *proto.FileSpec, r io.Reader) (outErr error) {
			defer utils.CatchAsError(&outErr)

			// log.Println(`写文章附件：`, post.Id, spec.Path)
			utils.Must(tw.WriteHeader(&tar.Header{
				Typeflag: tar.TypeReg,
				Name:     spec.Path,
				Size:     int64(spec.Size),
				Mode:     int64(spec.Mode),
				ModTime:  time.Unix(int64(spec.Time), 0),
			}))

			utils.Must1(io.Copy(tw, r))

			fileCount++

			return nil
		}))

		utils.Must(tw.Close())
		r := utils.Must1(wc.Close())

		if fileCount > 0 {
			f := fmt.Sprintf(`posts.%d.tar.gz.age`, post.Id)
			log.Println(`正在写入备份文件：`, f)
			utils.Must(b.remote.Upload(ctx, f, r, ``))
		}
	}

	b.store.SetInteger(`last_time`, now.Unix())

	return nil
}
