package main

import (
	"os"
	"path/filepath"
)

// FileRedirect does file redirection.
type FileRedirect struct {
	dir  string // the files directory
	host string // the file host to be used as alternative
}

// NewFileRedirect creates a new file redirect-or.
func NewFileRedirect(dir string, host string) *FileRedirect {
	return &FileRedirect{
		dir:  dir,
		host: host,
	}
}

// Redirect returns the redirected file location.
func (z *FileRedirect) Redirect(loggedin bool, file string) string {
	relative := filepath.Join("/", z.dir, file)
	local := "." + relative

	if loggedin {
		if _, err := os.Stat(local); err == nil {
			return relative
		}
	}

	// < file does not exist
	return z.host + relative
}
