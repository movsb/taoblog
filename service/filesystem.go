package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	fspkg "io/fs"
	"log"
	"os"
	"time"

	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) FileSystem(srv protocols.Management_FileSystemServer) error {
	if !s.auth.AuthGRPC(srv.Context()).IsAdmin() {
		return status.Error(codes.Unauthenticated, "bad credentials")
	}

	initialized := false

	var fs _FileSystem

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
				fs, err = s.fileSystemForPost(init.Id)
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

type _FileSystem interface {
	ListFiles() ([]*protocols.FileSpec, error)
	DeleteFile(path string) error
	WriteFile(spec *protocols.FileSpec, r io.Reader) error
}

type _FileSystemForPost struct {
	s  *Service
	id int64
}

func (fs *_FileSystemForPost) ListFiles() ([]*protocols.FileSpec, error) {
	filePaths, err := fs.s.Store().List(fs.id)
	if err != nil {
		return nil, err
	}

	files := []*protocols.FileSpec{}
	for _, path := range filePaths {
		spec, err := func() (*protocols.FileSpec, error) {
			fp, err := fs.s.Store().Open(fs.id, path)
			if err != nil {
				return nil, err
			}
			defer fp.Close()
			stat, err := fp.Stat()
			if err != nil {
				return nil, err
			}
			return &protocols.FileSpec{
				Path: path,
				Mode: uint32(stat.Mode()),
				Size: uint32(stat.Size()),
				Time: uint32(stat.ModTime().Unix()),
			}, nil
		}()
		if err != nil {
			return nil, err
		}
		files = append(files, spec)
	}

	return files, nil
}

func (fs *_FileSystemForPost) DeleteFile(path string) error {
	return fs.s.Store().Remove(fs.id, path)
}

func (fs *_FileSystemForPost) WriteFile(spec *protocols.FileSpec, r io.Reader) error {
	tmp, err := os.CreateTemp("", `taoblog-*`)
	if err != nil {
		return err
	}

	if n, err := io.Copy(tmp, r); err != nil || n != int64(spec.Size) {
		return fmt.Errorf(`write error: %d %v`, n, err)
	}

	if err := tmp.Chmod(fspkg.FileMode(spec.Mode)); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	t := time.Unix(int64(spec.Time), 0)

	if err := os.Chtimes(tmp.Name(), t, t); err != nil {
		return err
	}

	path, _ := fs.s.Store().PathOf(fs.id, spec.Path)
	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}

	return nil
}

func (s *Service) fileSystemForPost(id int64) (*_FileSystemForPost, error) {
	_ = s.MustGetPost(id)

	return &_FileSystemForPost{
		s:  s,
		id: id,
	}, nil
}
