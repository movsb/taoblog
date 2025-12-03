package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/storage"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taoblog/theme/modules/handle304"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) FileSystem(srv proto.Management_FileSystemServer) (outErr error) {
	defer utils.CatchAsError(&outErr)

	user.MustNotBeGuest(srv.Context())

	initialized := false

	var pfs fs.FS

	for {
		req, err := srv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			if st, ok := status.FromError(err); ok && st.Code() == codes.Canceled {
				return nil
			}
			log.Println(`接收消息失败：`, err)
			return err
		}

		initReq := req.GetInit()
		if !initialized && initReq == nil {
			log.Println(`没收到初始化消息。`)
			return status.Error(codes.FailedPrecondition, "not init")
		} else if initialized && initReq != nil {
			log.Println(`重复初始化。`)
			return status.Error(codes.Aborted, "re-init")
		} else if initReq != nil {
			initialized = true
			if init := initReq.GetPost(); init != nil {
				// TODO 没鉴权。
				pfs = s.postDataFS.ForPost(int(init.Id))
				err = nil
			}
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			if pfs == nil {
				return status.Error(codes.InvalidArgument, "unknown file system to operate")
			}
			if err := srv.Send(&proto.FileSystemResponse{
				Init: &proto.FileSystemResponse_InitResponse{},
			}); err != nil {
				return err
			}
			continue
		}

		if pfs == nil {
			return status.Error(codes.Internal, "not init")
		}

		if list := req.GetListFiles(); list != nil {
			files, err := utils.ListFiles(pfs)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					return err
				}
			}
			if err = srv.Send(&proto.FileSystemResponse{
				Response: &proto.FileSystemResponse_ListFiles{
					ListFiles: &proto.FileSystemResponse_ListFilesResponse{
						Files: files,
					},
				},
			}); err != nil {
				return err
			}
		} else if write := req.GetWriteFile(); write != nil {
			if err := utils.Write(pfs, write.Spec, bytes.NewReader(write.Data)); err != nil {
				log.Println(err)
				return err
			}
			if err = srv.Send(&proto.FileSystemResponse{
				Response: &proto.FileSystemResponse_WriteFile{
					WriteFile: &proto.FileSystemResponse_WriteFileResponse{},
				},
			}); err != nil {
				log.Println(err)
				return err
			}
		} else if delete := req.GetDeleteFile(); delete != nil {
			if err := utils.Delete(pfs, delete.Path); err != nil {
				return err
			}
			if err = srv.Send(&proto.FileSystemResponse{
				Response: &proto.FileSystemResponse_DeleteFile{
					DeleteFile: &proto.FileSystemResponse_DeleteFileResponse{},
				},
			}); err != nil {
				return err
			}
		} else if read := req.GetReadFile(); read != nil {
			r := utils.Must1(pfs.Open(read.Path))
			utils.Must(srv.Send(&proto.FileSystemResponse{
				Response: &proto.FileSystemResponse_ReadFile{
					ReadFile: &proto.FileSystemResponse_ReadFileResponse{
						Data: utils.Must1(io.ReadAll(r)),
					},
				},
			}))
		}
	}
}

func (s *Service) DeletePostFile(ctx context.Context, in *proto.DeletePostFileRequest) (_ *proto.DeletePostFileResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)
	po := utils.Must1(s.getPostCached(ctx, int(in.PostId)))
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || ac.User.ID == int64(po.UserID)) {
		panic(noPerm)
	}

	pfs := s.postDataFS.ForPost(int(in.PostId))
	if err := utils.Delete(pfs, in.Path); err != nil {
		if os.IsNotExist(err) {
			panic(status.Error(codes.NotFound, `file not found`))
		}
		panic(err)
	}

	return &proto.DeletePostFileResponse{}, nil
}

func (s *Service) UpdateFileCaption(ctx context.Context, in *proto.UpdateFileCaptionRequest) (_ *proto.UpdateFileCaptionResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)
	po := utils.Must1(s.getPostCached(ctx, int(in.PostId)))
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || ac.User.ID == int64(po.UserID)) {
		panic(noPerm)
	}

	pfs, ok := s.postDataFS.ForPost(int(in.PostId)).(*storage.SQLiteForPost)
	if !ok {
		return nil, status.Error(codes.Unimplemented, `此文件系统不支持更新文件元数据。`)
	}

	utils.Must(pfs.UpdateCaption(in.Path, &proto.FileSpec_Meta_Source{
		Format:  proto.FileSpec_Meta_Source_Markdown,
		Caption: in.Caption,
	}))

	s.deletePostContentCacheFor(int64(in.PostId))
	s.updatePostMetadataTime(int64(in.PostId), time.Now())

	return &proto.UpdateFileCaptionResponse{}, nil
}

const noPerm = `此操作无权限。`

func (s *Service) ListPostFiles(ctx context.Context, in *proto.ListPostFilesRequest) (_ *proto.ListPostFilesResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)
	po := utils.Must1(s.getPostCached(ctx, int(in.PostId)))
	if !(ac.User.IsSystem() || ac.User.ID == int64(po.UserID)) {
		return nil, status.Error(codes.PermissionDenied, noPerm)
	}

	pfs := s.postDataFS.ForPost(int(in.PostId))
	files := utils.Must1(utils.ListFiles(pfs))

	filtered := make([]*proto.FileSpec, 0, len(files))
	for _, file := range files {
		// 过滤自动生成的文件。
		if !in.WithGenerated && file.ParentPath != `` {
			continue
		}

		// 过滤实况照片视频？
		// 低效的写法，但是鉴于文件少，可行。
		if !in.WithLivePhotoVideo {
			ext := path.Ext(file.Path)
			if strings.HasPrefix(mime.TypeByExtension(ext), `video/`) {
				base1 := strings.TrimSuffix(file.Path, ext)
				// 依次判断有没有同名（不含后缀）的图片文件。
				found := false
				for _, file := range files {
					ext := path.Ext(file.Path)
					if strings.HasPrefix(mime.TypeByExtension(ext), `image/`) {
						base2 := strings.TrimSuffix(file.Path, ext)
						if base2 == base1 {
							found = true
							break
						}
					}
				}
				if found {
					continue
				}
			}
		}

		filtered = append(filtered, file)
	}

	return &proto.ListPostFilesResponse{
		Files: filtered,
	}, nil
}

func (s *Service) RegisterFileURLGetter(name string, g theme_fs.FileURLGetter) {
	s.fileURLGetters.Store(name, g)
}

func (s *Service) ServeFile(w http.ResponseWriter, r *http.Request, postID int64, file string, localOnly bool) {
	// 此处没有鉴权。
	p := utils.Must1(s.getPostCached(r.Context(), int(postID)))

	// 此处鉴权。
	// 不调用 GetPost，并发多图片下性能不好。
	ac := user.Context(r.Context())
	switch {
	// 系统有所有权限
	case ac.User.IsSystem():
		break
	// 是公开文章
	case p.Status == models.PostStatusPublic:
		break
	// 是本人
	case ac.User.ID == int64(p.UserID):
		break
	// 分享文章。
	case p.Status == models.PostStatusPartial && s.canNonAuthorUserReadPost(r.Context(), ac.User.ID, int(postID)):
		break
	default:
		http.NotFound(w, r)
		return
	}

	// 仅系统和本人可以访问特殊文件：以 . 或者 _ 开头的文件或目录。
	// 注意：分享用户也无法访问。
	if !(ac.User.IsSystem() || ac.User.ID == int64(p.UserID)) {
		for part := range strings.SplitSeq(file, `/`) {
			if part[0] == '.' || part[0] == '_' {
				http.Error(w, `尝试访问不允许访问的文件。`, http.StatusForbidden)
				return
			}
		}
	}

	pfs := s.postDataFS.ForPost(int(postID))

	if !localOnly {
		if cache := s.getFasterFileURL(r, pfs, p, file); cache != nil && cache.Get != `` {
			w.Header().Add(`Cache-Control`, `no-store`)
			if !cache.Encrypted {
				http.Redirect(w, r, cache.Get, http.StatusFound)
				return
			} else {
				// 对于加密情况，直接写相关加密参数，onerror 会处理。
				// 这样可以尽量减少请求导致的流量，加快速度。
				w.Header().Set(`Content-Type`, `application/json`)
				json.NewEncoder(w).Encode(map[string]any{
					`src`:   cache.Get,
					`key`:   base64.StdEncoding.EncodeToString(cache.Key),
					`nonce`: base64.StdEncoding.EncodeToString(cache.Nonce),

					`content_type`: mime.TypeByExtension(path.Ext(file)),
				})
				return
			}
		}
	}

	handle304.MustRevalidate(w)
	http.ServeFileFS(w, r, pfs, file)
}

type _FileURLCacheKey struct {
	Pid    int
	Status string
	Path   string
	China  bool
}

type _FileURLCacheValue struct {
	Time time.Time

	Get  string
	Head string

	Encrypted bool
	Nonce     []byte
	Key       []byte
}

func (s *Service) getFasterFileURL(r *http.Request, pfs fs.FS, p *models.Post, file string) *_FileURLCacheValue {
	getterCount := 0
	// 神经，居然不能获取大小。
	s.fileURLGetters.Range(func(key, value any) bool {
		getterCount++
		return true
	})
	if getterCount <= 0 {
		return nil
	}

	ac := user.Context(r.Context())

	const ttl = time.Hour * 24

	key := _FileURLCacheKey{
		Pid:    int(p.ID),
		Status: p.Status,
		Path:   file,
		China:  ac.InChina,
	}

	if val, ok := s.fileURLs.Peek(key); ok && time.Since(val.Time) < ttl-time.Minute {
		// 更改权限会导致文件立即失效，所有总是校验。
		rsp, err := http.Head(val.Head)
		if err == nil {
			defer rsp.Body.Close()
			if rsp.StatusCode == 200 {
				return val
			}
		}
	}

	s.fileURLs.Delete(key)

	val, err, _ := s.fileURLs.GetOrLoad(r.Context(), key, func(ctx context.Context, fuk _FileURLCacheKey) (*_FileURLCacheValue, error) {
		fp, err := pfs.Open(file)
		if err != nil {
			return nil, err
		}
		defer fp.Close()
		info := utils.Must1(fp.Stat())
		file, ok := info.Sys().(*models.File)
		if !ok {
			return nil, fmt.Errorf(`not user uploaded file`)
		}

		// 垃圾阿里云会自动给 html 文件增加 `Content-Disposition: attachment`，导致变成下载。
		//
		// "使用 OSS 默认域名访问 html、图片资源，会有以附件形式下载的情况。若需要浏览器直接访问，需使用自定义域名进行访问，了解详情。"
		//
		// https://help.aliyun.com/zh/oss/user-guide/how-to-ensure-an-object-is-previewed-when-you-access-the-object
		//
		// 在此特殊处理一下：
		//
		//  1. 如果是小文件，不走加速。
		if info.Size() < 100<<10 {
			return nil, io.EOF
		}

		var chinaVal, otherVal *_FileURLCacheValue

		postIsPublic := p.Status == models.PostStatusPublic

		// 文件摘要。
		remoteFileDigest := utils.IIF(postIsPublic,
			file.Digest,
			file.Meta.Encryption.Digest,
		)
		// 文件在远程服务器的路径。
		remotePath := utils.IIF(postIsPublic,
			path.Join(`files`, fmt.Sprint(file.PostID), file.Path),
			path.Join(`objects`, fmt.Sprint(file.PostID), remoteFileDigest),
		)

		s.fileURLGetters.Range(func(_, value any) bool {
			val := &_FileURLCacheValue{
				Time: time.Now(),
			}
			getter := value.(theme_fs.FileURLGetter)
			if get, head, err := getter.GetFileURL(remotePath, remoteFileDigest, ttl); err == nil {
				val.Get = get
				val.Head = head
				val.Encrypted = !postIsPublic
				if val.Encrypted {
					val.Nonce = file.Meta.Encryption.Nonce
					val.Key = file.Meta.Encryption.Key
				}

				if getter.GetCountry() == `china` {
					chinaVal = val
				} else {
					otherVal = val
				}
			}
			return true
		})

		// 用户在中国，且在中国有存储，则使用。
		if key.China && chinaVal != nil {
			return chinaVal, nil
		}

		// 用户不在中国或者没在中国存储，则使用外国的。
		if otherVal != nil {
			return otherVal, nil
		}
		// 外国没配，那就再使用中国的。
		if chinaVal != nil {
			return chinaVal, nil
		}

		return nil, io.EOF
	})

	if err != nil {
		return nil
	}

	return val
}
