package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/modules/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) FileSystem(srv protocols.Management_FileSystemServer) error {
	// TODO 如果是评论，允许用户上传文件。
	s.MustBeAdmin(srv.Context())

	initialized := false

	var fs storage.FileSystem

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
				fs, err = s.FileSystemForPost(srv.Context(), init.Id)
			}
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			if fs == nil {
				return status.Error(codes.InvalidArgument, "unknown file system to operate")
			}
			if err := srv.Send(&protocols.FileSystemResponse{
				Init: &protocols.FileSystemResponse_InitResponse{},
			}); err != nil {
				return err
			}
			continue
		}

		if fs == nil {
			return status.Error(codes.Internal, "not init")
		}

		if list := req.GetListFiles(); list != nil {
			files, err := fs.ListFiles()
			if err != nil {
				return err
			}
			if err = srv.Send(&protocols.FileSystemResponse{
				Response: &protocols.FileSystemResponse_ListFiles{
					ListFiles: &protocols.FileSystemResponse_ListFilesResponse{
						Files: files,
					},
				},
			}); err != nil {
				return err
			}
		} else if write := req.GetWriteFile(); write != nil {
			if err := fs.WriteFile(write.Spec, bytes.NewReader(write.Data)); err != nil {
				log.Println(err)
				return err
			}
			if err = srv.Send(&protocols.FileSystemResponse{
				Response: &protocols.FileSystemResponse_WriteFile{
					WriteFile: &protocols.FileSystemResponse_WriteFileResponse{},
				},
			}); err != nil {
				log.Println(err)
				return err
			}
		} else if delete := req.GetDeleteFile(); delete != nil {
			if err := fs.DeleteFile(delete.Path); err != nil {
				return err
			}
			if err = srv.Send(&protocols.FileSystemResponse{
				Response: &protocols.FileSystemResponse_DeleteFile{
					DeleteFile: &protocols.FileSystemResponse_DeleteFileResponse{},
				},
			}); err != nil {
				return err
			}
		}
	}
}

func (s *Service) FileSystemForPost(ctx context.Context, id int64) (*storage.Local, error) {
	if s.testing {
		panic(`测试服务器不用于本地文件系统。`)
	}
	// _ = s.MustGetPost(id)
	maxFileSize := int32(1 << 20)
	if ac := auth.Context(ctx); ac != nil && ac.User.IsAdmin() {
		maxFileSize = 100 << 20
	}
	return storage.NewLocal(s.cfg.Data.File.Path, id,
		storage.WithMaxFileSize(maxFileSize),
	), nil
}
