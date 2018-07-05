package main

import (
	"os"
	"path/filepath"
)

// FileRedirect does file redirection.
type FileRedirect struct {
	root string // the taoblog root path
	dir  string // the files directory
	host string // the file host to be used as alternative
}

// NewFileRedirect creates a new file redirect-or.
func NewFileRedirect(root string, dir string, host string) *FileRedirect {
	return &FileRedirect{
		root: root,
		dir:  dir,
		host: host,
	}
}

// Redirect returns the redirected file location.
func (z *FileRedirect) Redirect(loggedin bool, file string) string {
	relative := filepath.Join("/", z.dir, file)
	absolute := z.root + relative

	if loggedin {
		if _, err := os.Stat(absolute); err == nil {
			return relative
		}
	}

	// < file does not exist
	return z.host + relative
}
