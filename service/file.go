package service

import (
	"fmt"
	"io"
	"path/filepath"
)

func filePath(postID int64, file string) string {
	path := filepath.Clean(filepath.Join("/", fmt.Sprint(postID), file))
	return filepath.Join("./files", path)
}

func (s *ImplServer) GetFile(postID int64, file string) string {
	return filePath(postID, file)
}

func (s *ImplServer) UploadFile(postID int64, file string, data io.Reader) error {
	return s.fmgr.Put(postID, file, data)
}

func (s *ImplServer) ListFiles(postID int64) ([]string, error) {
	return s.fmgr.List(postID)
}

func (s *ImplServer) DeleteFile(postID int64, file string) error {
	return s.fmgr.Delete(postID, file)
}
