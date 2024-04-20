package service

import (
	"bytes"
	"errors"
	"io"
	"log"

	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/modules/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) FileSystem(srv protocols.Management_FileSystemServer) error {
	if !s.auth.AuthGRPC(srv.Context()).IsAdmin() {
		return status.Error(codes.Unauthenticated, "bad credentials")
	}

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
				fs, err = s.FileSystemForPost(init.Id)
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

func (s *Service) FileSystemForPost(id int64) (*storage.Local, error) {
	// _ = s.MustGetPost(id)
	return storage.NewLocal(s.cfg.Data.File.Path, id), nil
}
