package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/auth"
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

	// TODO 如果是评论，允许用户上传文件。
	auth.MustNotBeGuest(srv.Context())

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
				pfs, err = s.postDataFS.ForPost(int(init.Id))
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

	ac := auth.MustNotBeGuest(ctx)
	po := utils.Must1(s.getPostCached(ctx, int(in.PostId)))
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || ac.User.ID == int64(po.UserID)) {
		panic(noPerm)
	}

	pfs := utils.Must1(s.postDataFS.ForPost(int(in.PostId)))
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

	ac := auth.MustNotBeGuest(ctx)
	po := utils.Must1(s.getPostCached(ctx, int(in.PostId)))
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || ac.User.ID == int64(po.UserID)) {
		panic(noPerm)
	}

	// TODO: 强制转换了，只应有这一种实例。
	pfs := utils.Must1(s.postDataFS.ForPost(int(in.PostId))).(*storage.SQLiteForPost)
	utils.Must(pfs.UpdateCaption(in.Path, &proto.FileSpec_Meta_Source{
		Format:  proto.FileSpec_Meta_Source_Markdown,
		Caption: in.Caption,
	}))

	s.deletePostContentCacheFor(int64(in.PostId))
	s.updatePostMetadataTime(int64(in.PostId), time.Now())

	return &proto.UpdateFileCaptionResponse{}, nil
}

func (s *Service) ListPostFiles(ctx context.Context, in *proto.ListPostFilesRequest) (_ *proto.ListPostFilesResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := auth.MustNotBeGuest(ctx)
	po := utils.Must1(s.getPostCached(ctx, int(in.PostId)))
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || ac.User.ID == int64(po.UserID)) {
		panic(noPerm)
	}

	pfs := utils.Must1(s.postDataFS.ForPost(int(in.PostId)))
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
	// 权限检查
	// TODO 应该调用带缓存接口（注意鉴权），否则会频繁查数据库。
	p := utils.Must1(s.GetPost(r.Context(), &proto.GetPostRequest{Id: int32(postID)}))

	// 所有人禁止访问特殊文件：以 . 或者 _ 开头的文件或目录。
	// TODO：以及 config.yaml | README.md
	switch file[0] {
	case '.', '_':
		panic(status.Error(codes.PermissionDenied, `尝试访问不允许访问的文件。`))
	}
	switch path.Base(file)[0] {
	case '.', '_':
		panic(status.Error(codes.PermissionDenied, `尝试访问不允许访问的文件。`))
	}
	// 为了不区分大小写，所以没有用 switch。
	if strings.EqualFold(file, `config.yml`) || strings.EqualFold(file, `config.yaml`) || strings.EqualFold(file, `README.md`) {
		panic(status.Error(codes.PermissionDenied, `尝试访问不允许访问的文件。`))
	}

	pfs := utils.Must1(s.postDataFS.ForPost(int(postID)))

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
}

type _FileURLCacheValue struct {
	Get       string
	Head      string
	Encrypted bool
	Time      time.Time
	Nonce     []byte
	Key       []byte
}

func (s *Service) getFasterFileURL(r *http.Request, pfs fs.FS, p *proto.Post, file string) *_FileURLCacheValue {
	getterCount := 0
	// 神经，居然不能获取大小。
	s.fileURLGetters.Range(func(key, value any) bool {
		getterCount++
		return true
	})
	if getterCount <= 0 {
		return nil
	}

	ac := auth.Context(r.Context())
	// 测试阶段，只给登录用户使用。
	if ac.User.IsGuest() {
		return nil
	}

	const ttl = time.Hour

	key := _FileURLCacheKey{Pid: int(p.Id), Status: p.Status, Path: file}
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
		val := _FileURLCacheValue{
			Time: time.Now(),
		}
		fp, err := pfs.Open(file)
		if err != nil {
			return nil, err
		}
		defer fp.Close()
		info := utils.Must1(fp.Stat())
		file := info.Sys().(*models.File)

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

		s.fileURLGetters.Range(func(key, value any) bool {
			if get, head, enc, err := value.(theme_fs.FileURLGetter).GetFileURL(p, file, ttl); err == nil {
				val.Get = get
				val.Head = head
				val.Encrypted = enc
				if enc {
					val.Nonce = file.Meta.Encryption.Nonce
					val.Key = file.Meta.Encryption.Key
				}
				return false
			}
			return true
		})

		if val.Get == `` {
			return nil, io.EOF
		}

		return &val, nil
	})

	if err != nil {
		return nil
	}

	return val
}
