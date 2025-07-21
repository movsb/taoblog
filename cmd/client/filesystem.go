package client

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/movsb/taoblog/modules/utils/syncer"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type SyncFileSpec proto.FileSpec

func (s *SyncFileSpec) Compare(than *SyncFileSpec) int {
	return strings.Compare(s.Path, than.Path)
}

func (s *SyncFileSpec) DeepEqual(to *SyncFileSpec) bool {
	lm, rm := os.FileMode(s.Mode), os.FileMode(to.Mode)
	if lm.IsDir() != rm.IsDir() {
		return false
	}
	if s.Size != to.Size {
		return false
	}
	if s.Time != to.Time {
		return false
	}
	return true
}

type FilesSyncer struct {
	client proto.Management_FileSystemClient
}

func NewFilesSyncer(client proto.Management_FileSystemClient) *FilesSyncer {
	s := &FilesSyncer{
		client: client,
	}
	return s
}

func (s *FilesSyncer) SyncPostFiles(id int64, root fs.FS, files []string) error {
	localFiles, err := s.listLocalFilesFromPaths(root, files)
	if err != nil {
		return err
	}

	if err := s.init(&proto.FileSystemRequest_InitRequest{
		For: &proto.FileSystemRequest_InitRequest_Post_{
			Post: &proto.FileSystemRequest_InitRequest_Post{
				Id: id,
			},
		},
	}); err != nil {
		return err
	}

	wrappedLocalFiles := s.wrapFileSpecs(localFiles)
	wrappedRemoteFiles, err := s.listRemoteFiles()
	if err != nil {
		return err
	}

	ss := syncer.New(
		syncer.WithCopyLocalToRemote[[]*SyncFileSpec](func(f *SyncFileSpec) error {
			fp, err := root.Open(f.Path)
			if err != nil {
				return err
			}
			defer fp.Close()
			return s.copyLocalToRemote(f, fp)
		}),
	)

	return ss.Sync(wrappedLocalFiles, wrappedRemoteFiles, syncer.LocalToRemote)
}

func (s *FilesSyncer) init(req *proto.FileSystemRequest_InitRequest) error {
	if err := s.client.Send(&proto.FileSystemRequest{
		Init: req,
	}); err != nil {
		panic(err)
	}
	rsp, err := s.client.Recv()
	if err != nil {
		return err
	}
	if rsp.GetInit() == nil {
		return fmt.Errorf("expect init")
	}
	return nil
}

func (s *FilesSyncer) listLocalFilesFromPaths(root fs.FS, paths []string) ([]*proto.FileSpec, error) {
	var localFiles []*proto.FileSpec
	for _, file := range paths {
		stat, err := fs.Stat(root, file)
		if err != nil {
			return nil, err
		}
		f := proto.FileSpec{
			Path: file,
			Mode: uint32(stat.Mode()),
			Size: uint32(stat.Size()),
			Time: uint32(stat.ModTime().Unix()),
		}
		localFiles = append(localFiles, (*proto.FileSpec)(&f))
	}
	return localFiles, nil
}

func (s *FilesSyncer) listRemoteFiles() ([]*SyncFileSpec, error) {
	if err := s.client.Send(&proto.FileSystemRequest{
		Request: &proto.FileSystemRequest_ListFiles{
			ListFiles: &proto.FileSystemRequest_ListFilesRequest{},
		},
	}); err != nil {
		return nil, err
	}
	rsp, err := s.client.Recv()
	if err != nil {
		return nil, err
	}
	remoteList := rsp.GetListFiles()
	if remoteList == nil {
		return nil, fmt.Errorf("remote list is nil")
	}
	remoteFiles := remoteList.GetFiles()
	return s.wrapFileSpecs(remoteFiles), nil
}

func (s *FilesSyncer) wrapFileSpecs(list []*proto.FileSpec) []*SyncFileSpec {
	sfs := make([]*SyncFileSpec, 0, len(list))
	for _, l := range list {
		sfs = append(sfs, (*SyncFileSpec)(l))
	}
	return sfs
}

func (s *FilesSyncer) deleteRemote(f *SyncFileSpec) error {
	if err := s.client.Send(&proto.FileSystemRequest{
		Request: &proto.FileSystemRequest_DeleteFile{
			DeleteFile: &proto.FileSystemRequest_DeleteFileRequest{
				Path: f.Path,
			},
		},
	}); err != nil {
		panic(err)
	}
	rsp, err := s.client.Recv()
	if err != nil {
		panic(err)
	}
	if rsp.GetDeleteFile() == nil {
		panic("expect get delete")
	}
	log.Println("删除远程：", f.Path)
	return nil
}

func (s *FilesSyncer) copyLocalToRemote(f *SyncFileSpec, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err := s.client.Send(&proto.FileSystemRequest{
		Request: &proto.FileSystemRequest_WriteFile{
			WriteFile: &proto.FileSystemRequest_WriteFileRequest{
				Spec: (*proto.FileSpec)(f),
				Data: data,
			},
		},
	}); err != nil {
		return err
	}
	rsp, err := s.client.Recv()
	if err != nil {
		return err
	}
	if rsp.GetWriteFile() == nil {
		return fmt.Errorf("expect write file")
	}
	log.Println("复制到远程：", f.Path)
	return nil
}
