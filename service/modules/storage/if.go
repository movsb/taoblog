package storage

import (
	"io"
	"os"
)

// File ...
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
	Stat() (os.FileInfo, error)
}

// Store exposes interfaces to manage post files.
type Store interface {
	List(id int64) ([]string, error)
	// path is cleaned.
	Open(id int64, path string) (File, error)
	// path is cleaned.
	Create(id int64, path string) (File, error)
	// path is cleaned.
	Remove(id int64, path string) error
}
