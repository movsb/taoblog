package models

import (
	"io"
	"io/fs"
	"path"
	"time"
)

type File struct {
	ID        int
	CreatedAt int64
	UpdatedAt int64
	PostID    int
	Path      string
	Mode      uint32
	ModTime   int64
	Size      uint32
	Data      []byte
}

func (File) TableName() string {
	return `files`
}

func (f *File) FsFile(r io.Reader) fs.File {
	return &_FsFile{f: f, r: r}
}

func (f *File) InfoFile() fs.FileInfo {
	return &_InfoFile{f: f}
}

func (f *File) DirEntry() fs.DirEntry {
	return &_DirEntry{f: f}
}

type _FsFile struct {
	f *File
	r io.Reader
}

func (f *_FsFile) Stat() (fs.FileInfo, error) { return &_InfoFile{f.f}, nil }
func (f *_FsFile) Read(p []byte) (int, error) { return f.r.Read(p) }
func (f *_FsFile) Close() error               { return nil }

type _InfoFile struct{ f *File }

func (f *_InfoFile) Name() string       { return path.Base(f.f.Path) }
func (f *_InfoFile) Size() int64        { return int64(f.f.Size) }
func (f *_InfoFile) Mode() fs.FileMode  { return fs.FileMode(f.f.Mode) }
func (f *_InfoFile) ModTime() time.Time { return time.Unix(f.f.ModTime, 0).Local() }
func (f *_InfoFile) IsDir() bool        { return f.Mode().IsDir() }
func (f *_InfoFile) Sys() any           { return nil }

type _DirEntry struct{ f *File }

func (f *_DirEntry) Name() string               { return path.Base(f.f.Path) }
func (f *_DirEntry) IsDir() bool                { return (&_InfoFile{f: f.f}).IsDir() }
func (f *_DirEntry) Type() fs.FileMode          { return (&_InfoFile{f: f.f}).Mode().Type() }
func (f *_DirEntry) Info() (fs.FileInfo, error) { return &_InfoFile{f: f.f}, nil }
