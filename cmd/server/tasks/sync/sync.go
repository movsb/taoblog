package server_sync_tasks

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	pathpkg "path"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/backups/oss"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
)

type SyncToOSS struct {
	oss     oss.Client
	options utils.PluginStorage
	blog    proto.TaoBlogServer
	pfs     theme_fs.FS
}

func NewSyncToOSS(ctx context.Context, provider string, cfg *config.OSSConfig, server proto.TaoBlogServer, options utils.PluginStorage, pfs theme_fs.FS) (*SyncToOSS, error) {
	oss, err := oss.New(provider, cfg)
	if err != nil {
		return nil, err
	}

	sr2 := &SyncToOSS{
		oss:     oss,
		options: options,
		blog:    server,
		pfs:     pfs,
	}

	go sr2.Run(ctx)

	return sr2, nil
}

func (s *SyncToOSS) Run(ctx context.Context) {
	time.Sleep(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
			if err := s.run(ctx); err != nil {
				log.Println(err)
			}
		}
	}
}

// TODO 可选不上传私有文章。
func (s *SyncToOSS) run(ctx context.Context) (outErr error) {
	defer utils.CatchAsError(&outErr)
	last := utils.Must1(s.options.GetIntegerDefault(`last`, 1))
	now := time.Now()
	updated := utils.Must1(s.blog.ListPosts(
		auth.SystemForLocal(ctx),
		&proto.ListPostsRequest{
			ModifiedNotBefore: int32(last),
		},
	)).GetPosts()

	for _, up := range updated {
		pfs := s.pfs.ForPost(int(up.Id))
		specs := utils.Must1(utils.ListFiles(pfs))
		for _, spec := range specs {
			utils.Must(s.upload(ctx, up, pfs, int(up.Id), spec.Path))
		}
		// 删除不再需要的文件：
		// - 公开后，以前加密的
		// - 私密后，以前公开的
		if up.Status == models.PostStatusPublic {
			s.oss.DeleteByPrefix(ctx, fmt.Sprintf(`objects/%d/`, up.Id))
		} else {
			s.oss.DeleteByPrefix(ctx, fmt.Sprintf(`files/%d/`, up.Id))
		}
	}

	if len(updated) > 0 {
		log.Println(`Finished uploading all.`)
	}

	utils.Must(s.options.SetInteger(`last`, now.Unix()))
	return
}

func (s *SyncToOSS) upload(ctx context.Context, post *proto.Post, pfs fs.FS, pid int, path string) error {
	fp := utils.Must1(pfs.Open(path))
	defer fp.Close()
	info := utils.Must1(fp.Stat())
	sysFile := info.Sys().(*models.File)

	log.Println(`正在上传文件到对象存储:`, pid, path)

	var digest string
	var fullPath string
	var size int
	var reader io.Reader

	if post.Status == models.PostStatusPublic {
		// files/文章编号/文件路径，没有前缀 /。
		fullPath = pathpkg.Join(`files`, fmt.Sprint(pid), path)
		digest = sysFile.Digest
		size = int(info.Size())
		reader = fp
	} else {
		digest = sysFile.Meta.Encryption.Digest
		fullPath = fmt.Sprintf(`objects/%d/%s`, pid, digest)
		size = sysFile.Meta.Encryption.Size

		aes := utils.Must1(aes.NewCipher(sysFile.Meta.Encryption.Key))
		aead := utils.Must1(cipher.NewGCM(aes))
		data := make([]byte, 0, size)
		data = aead.Seal(data, sysFile.Meta.Encryption.Nonce, utils.Must1(io.ReadAll(fp)), nil)
		if len(data) != size {
			panic(`加密数据长度不一样`)
		}
		reader = bytes.NewReader(data)
	}

	if err := s.oss.Upload(
		ctx, fullPath, int64(size), reader,
		mime.TypeByExtension(pathpkg.Ext(fullPath)),
		oss.NewDigestFromString(digest),
	); err != nil {
		log.Println(`上传失败：`, fullPath, err)
		return err
	}

	return nil
}

func (s *SyncToOSS) GetFileURL(post *proto.Post, file *models.File, ttl time.Duration) (string, string, bool, error) {
	var (
		path      string
		digest    string
		encrypted bool
	)

	if post.Status == models.PostStatusPublic {
		digest = file.Digest
		path = pathpkg.Join(`files`, fmt.Sprint(post.Id), file.Path)
		encrypted = false
	} else {
		digest = file.Meta.Encryption.Digest
		path = pathpkg.Join(`objects`, fmt.Sprint(post.Id), digest)
		encrypted = true
	}

	get, head, err := s.oss.GetFileURL(context.Background(), path, oss.NewDigestFromString(digest), ttl)
	if err != nil {
		return ``, ``, false, err
	}
	return get, head, encrypted, nil
}
