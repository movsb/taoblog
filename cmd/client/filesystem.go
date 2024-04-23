package client

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/movsb/taoblog/modules/utils/syncer"
	"github.com/movsb/taoblog/protocols"
)

type SyncFileSpec protocols.FileSpec

func (s *SyncFileSpec) Less(than *SyncFileSpec) bool {
	return s.Path < than.Path
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
	client protocols.Management_FileSystemClient
}

func NewFilesSyncer(client protocols.Management_FileSystemClient) *FilesSyncer {
	s := &FilesSyncer{
		client: client,
	}
	return s
}

func (s *FilesSyncer) SyncPostFiles(id int64, localFiles []*protocols.FileSpec) error {
	if err := s.init(&protocols.FileSystemRequest_InitRequest{
		For: &protocols.FileSystemRequest_InitRequest_Post_{
			Post: &protocols.FileSystemRequest_InitRequest_Post{
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
			fp, err := os.Open(f.Path)
			if err != nil {
				return err
			}
			defer fp.Close()
			return s.copyLocalToRemote(f, fp)
		}),
	)

	return ss.Sync(wrappedLocalFiles, wrappedRemoteFiles, syncer.LocalToRemote)
}

func (s *FilesSyncer) init(req *protocols.FileSystemRequest_InitRequest) error {
	if err := s.client.Send(&protocols.FileSystemRequest{
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

func (s *FilesSyncer) ListLocalFilesFromPaths(paths []string) ([]*protocols.FileSpec, error) {
	var localFiles []*protocols.FileSpec
	for _, file := range paths {
		stat, err := os.Stat(file)
		if err != nil {
			return nil, err
		}
		f := protocols.FileSpec{
			Path: file,
			Mode: uint32(stat.Mode()),
			Size: uint32(stat.Size()),
			Time: uint32(stat.ModTime().Unix()),
		}
		localFiles = append(localFiles, (*protocols.FileSpec)(&f))
	}
	return localFiles, nil
}

func (s *FilesSyncer) listRemoteFiles() ([]*SyncFileSpec, error) {
	if err := s.client.Send(&protocols.FileSystemRequest{
		Request: &protocols.FileSystemRequest_ListFiles{
			ListFiles: &protocols.FileSystemRequest_ListFilesRequest{},
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

func (s *FilesSyncer) wrapFileSpecs(list []*protocols.FileSpec) []*SyncFileSpec {
	sfs := make([]*SyncFileSpec, 0, len(list))
	for _, l := range list {
		sfs = append(sfs, (*SyncFileSpec)(l))
	}
	return sfs
}

func (s *FilesSyncer) deleteRemote(f *SyncFileSpec) error {
	if err := s.client.Send(&protocols.FileSystemRequest{
		Request: &protocols.FileSystemRequest_DeleteFile{
			DeleteFile: &protocols.FileSystemRequest_DeleteFileRequest{
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
	if err := s.client.Send(&protocols.FileSystemRequest{
		Request: &protocols.FileSystemRequest_WriteFile{
			WriteFile: &protocols.FileSystemRequest_WriteFileRequest{
				Spec: (*protocols.FileSpec)(f),
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
