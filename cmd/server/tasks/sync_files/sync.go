package sync_files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	pathpkg "path"
	"strings"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/backups/oss"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/syncer"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
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
		case <-time.After(time.Second * 30):
			if err := s.run(ctx); err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *SyncToOSS) run(ctx context.Context) (outErr error) {
	defer utils.CatchAsError(&outErr)
	last := utils.Must1(s.options.GetIntegerDefault(`last`, 1))
	now := time.Now()
	updated := utils.Must1(s.blog.ListPosts(
		user.SystemForLocal(ctx),
		&proto.ListPostsRequest{
			ModifiedNotBefore: int32(last),
		},
	)).GetPosts()

	for _, post := range updated {
		pfs := s.pfs.ForPost(int(post.Id))
		specs := utils.Must1(utils.ListFiles(pfs))
		existed, err := s.oss.ListFiles(ctx, fmt.Sprintf(`objects/%d/`, post.Id))
		if err != nil {
			log.Println(`列出文件失败：`, err)
			return err
		}
		if err := s.syncPostFiles(ctx, post, pfs, specs, existed); err != nil {
			log.Println(`同步文章文件失败：`, post.Id, err)
			return err
		}
	}

	if len(updated) > 0 {
		log.Println(`Finished uploading all.`)
	}

	utils.Must(s.options.SetInteger(`last`, now.Unix()))
	return
}

type SyncerFileMeta struct {
	PostFilePath string
	oss.FileMeta
}

func (m SyncerFileMeta) Compare(other SyncerFileMeta) int {
	return strings.Compare(m.Path, other.Path)
}
func (m SyncerFileMeta) DeepEqual(other SyncerFileMeta) bool {
	return m.Digest.Equals(other.Digest)
}

func (s *SyncToOSS) syncPostFiles(ctx context.Context, post *proto.Post, pfs fs.FS, specs []*proto.FileSpec, existed []oss.FileMeta) error {
	newFiles := utils.Map(specs, func(spec *proto.FileSpec) SyncerFileMeta {
		digest := oss.NewDigestFromString(spec.Digest)
		return SyncerFileMeta{
			PostFilePath: spec.Path,
			FileMeta: oss.FileMeta{
				Path: fmt.Sprintf(`objects/%d/%s`, post.Id, digest.String()),
				// 这里始终是文件本身的摘要（而非加密后的）。
				Digest: digest,
			},
		}
	})

	oldFiles := utils.Map(existed, func(meta oss.FileMeta) SyncerFileMeta {
		return SyncerFileMeta{
			PostFilePath: ``, // 远程文件没有原始路径了。
			FileMeta:     meta,
		}
	})

	sync := syncer.New(
		syncer.WithCopyLocalToRemote[[]SyncerFileMeta](func(f SyncerFileMeta) error {
			log.Println(`上传文件到远程：`, f.PostFilePath, f.Digest)

			fp := utils.Must1(pfs.Open(f.PostFilePath))
			defer fp.Close()

			info := utils.Must1(fp.Stat())
			sysFile, ok := info.Sys().(*models.File)
			// 不是用户上传的普通文件。
			if !ok {
				return nil
			}

			var digest string
			var size int
			var reader io.Reader

			// NOTE: 公开文章的私有文件（以 _ 和 . 开头的那些）也应该加密保存。
			if post.Status == models.PostStatusPublic && !strings.HasPrefix(f.PostFilePath, `_`) && !strings.HasPrefix(f.PostFilePath, `.`) {
				digest = sysFile.Digest
				size = int(info.Size())
				reader = fp
			} else {
				digest = sysFile.Meta.Encryption.Digest
				size = sysFile.Meta.Encryption.Size
				encrypted := sysFile.Meta.Encryption.EncryptData(utils.Must1(io.ReadAll(fp)))
				if len(encrypted) != size {
					panic(`加密数据长度不一样`)
				}
				reader = bytes.NewReader(encrypted)
			}

			// TODO: 如果不同目录有相同文件，会出错。
			if err := s.oss.Upload(ctx,
				f.Path,
				int64(size), reader,
				mime.TypeByExtension(pathpkg.Ext(f.PostFilePath)),
				oss.NewDigestFromString(digest),
			); err != nil {
				log.Println(`上传失败：`, f.PostFilePath, f.Path, err)
				return err
			}
			return nil
		}),
		syncer.WithDeleteRemote[[]SyncerFileMeta](func(f SyncerFileMeta) error {
			log.Println(`删除远程文件：`, f.Path, f.Digest)
			// 注意：这里的 f.Path 是对象存储返回的路径，≠ 原始文件名。
			return s.oss.DeleteFile(ctx, f.Path)
		}),
	)

	return sync.Sync(newFiles, oldFiles, syncer.LocalToRemote)
}

func (s *SyncToOSS) GetFileURL(publicPost bool, file *models.File, ttl time.Duration) (string, string, bool, error) {
	path := fmt.Sprintf(`objects/%d/%s`, file.PostID, file.Digest)
	plain := publicPost && !strings.HasPrefix(file.Path, `_`) && !strings.HasPrefix(file.Path, `.`)
	digest := utils.IIF(plain, file.Digest, file.Meta.Encryption.Digest)
	get, head, err := s.oss.GetFileURL(context.Background(), path, oss.NewDigestFromString(digest), ttl)
	return get, head, !plain, err
}
