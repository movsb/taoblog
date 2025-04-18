package server_sync_tasks

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"mime"
	pathpkg "path"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/backups/r2"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
)

type SyncToR2 struct {
	r2      *r2.R2
	options utils.PluginStorage
	blog    proto.TaoBlogServer
	pfs     theme_fs.FS
}

func NewSyncToR2(ctx context.Context, cfg config.SiteSyncR2Config, server proto.TaoBlogServer, options utils.PluginStorage, pfs theme_fs.FS) (*SyncToR2, error) {
	r2, err := r2.New(cfg.AccountID, cfg.AccessKeyID, cfg.AccessKeySecret, cfg.BucketName)
	if err != nil {
		return nil, err
	}

	sr2 := &SyncToR2{
		r2:      r2,
		options: options,
		blog:    server,
		pfs:     pfs,
	}

	go sr2.Run(ctx)

	return sr2, nil
}

func (s *SyncToR2) Run(ctx context.Context) {
	time.Sleep(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute * 1):
			if err := s.run(ctx); err != nil {
				log.Println(err)
			}
		}
	}
}

// TODO 可选不上传私有文章。
func (s *SyncToR2) run(ctx context.Context) (outErr error) {
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
		pfs := utils.Must1(s.pfs.ForPost(int(up.Id)))
		specs := utils.Must1(utils.ListFiles(pfs))
		for _, spec := range specs {
			utils.Must(s.upload(ctx, pfs, int(up.Id), spec.Path))
		}
	}

	utils.Must(s.options.SetInteger(`last`, now.Unix()))
	return
}

func (s *SyncToR2) upload(ctx context.Context, pfs fs.FS, pid int, path string) error {
	fp := utils.Must1(pfs.Open(path))
	defer fp.Close()
	// files/文章编号/文件路径，没有前缀 /。
	fullPath := pathpkg.Join(`files`, fmt.Sprint(pid), path)
	log.Println(`正在上传文件到 R2:`, fullPath)
	if err := s.r2.Upload(ctx, fullPath, fp, mime.TypeByExtension(pathpkg.Ext(fullPath))); err != nil {
		log.Println(`上传失败：`, fullPath, err)
		return err
	}
	return nil
}

func (s *SyncToR2) GetFileURL(pid int, path string, digest string) string {
	fullPath := pathpkg.Join(`files`, fmt.Sprint(pid), path)
	return s.r2.GetFileURL(context.Background(), fullPath, digest)
}
