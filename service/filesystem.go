package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
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

func (s *Service) ListPostFiles(ctx context.Context, in *proto.ListPostFilesRequest) (_ *proto.ListPostFilesResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := auth.MustNotBeGuest(ctx)
	po := utils.Must1(s.getPost(int(in.PostId)))
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || ac.User.ID == int64(po.UserID)) {
		panic(noPerm)
	}

	pfs := utils.Must1(s.postDataFS.ForPost(int(in.PostId)))
	files := utils.Must1(utils.ListFiles(pfs))
	return &proto.ListPostFilesResponse{
		Files: files,
	}, nil
}

func (s *Service) RegisterFileURLGetter(name string, g theme_fs.FileURLGetter) {
	s.fileURLGetters.Store(name, g)
}

func (s *Service) HandlePostFile(w http.ResponseWriter, r *http.Request, pid int, path string) {
	handled := false

	pfs := utils.Must1(s.postDataFS.ForPost(pid))

	ac := auth.Context(r.Context())
	// 测试阶段，只给登录用户使用。
	if !ac.User.IsGuest() {
		key := _FileURLCacheKey{Pid: pid, Path: path}
		if val, ok := s.fileURLs.Peek(key); ok && time.Since(val.Time) < time.Minute*10 {
			http.Redirect(w, r, val.URL, http.StatusFound)
			return
		}

		s.fileURLs.Delete(key)

		val, err, _ := s.fileURLs.GetOrLoad(r.Context(), key, func(ctx context.Context, fuk _FileURLCacheKey) (_FileURLCacheValue, error) {
			fp, err := pfs.Open(path)
			if err != nil {
				return _FileURLCacheValue{}, err
			}
			defer fp.Close()
			info := utils.Must1(fp.Stat())
			file := info.Sys().(*models.File)

			var url string

			s.fileURLGetters.Range(func(key, value any) bool {
				if u := value.(theme_fs.FileURLGetter).GetFileURL(pid, path, file.Digest); u != `` {
					url = u
					return false
				}
				return true
			})

			if url == `` {
				return _FileURLCacheValue{}, io.EOF
			}

			return _FileURLCacheValue{URL: url, Time: time.Now()}, nil
		})

		if err == nil {
			http.Redirect(w, r, val.URL, http.StatusFound)
			return
		}
	}

	if !handled {
		http.ServeFileFS(w, r, pfs, path)
	}
}

type _FileURLCacheKey struct {
	Pid  int
	Path string
}

type _FileURLCacheValue struct {
	URL  string
	Time time.Time
}
