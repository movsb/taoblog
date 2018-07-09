package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IFileManager exposes interfaces to manage upload files.
type IFileManager interface {
	Put(pid int64, name string, r io.Reader) error
	Delete(pid int64, name string) error
	List(pid int64) ([]string, error)
}

// FileUpload operates on upload files
type FileUpload struct {
	mgr IFileManager
}

// NewFileUpload returns a new instance of FileUpload
func NewFileUpload(mgr IFileManager) *FileUpload {
	return &FileUpload{
		mgr: mgr,
	}
}

// Upload does file saving
func (o *FileUpload) Upload(c *gin.Context) error {
	var err error

	parent := toInt64(c.Param("parent"))

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	for _, file := range form.File["files[]"] {
		fp, err := file.Open()
		if err != nil {
			return err
		}
		defer fp.Close()
		if err = o.mgr.Put(parent, file.Filename, fp); err != nil {
			return err
		}
	}

	return nil
}

// List does file listing
func (o *FileUpload) List(c *gin.Context) ([]string, error) {
	parent := toInt64(c.Param("parent"))
	return o.mgr.List(parent)
}

// Delete does file deleting
func (o *FileUpload) Delete(c *gin.Context) error {
	pidstr, ok := c.GetPostForm("pid")
	pid, err := strconv.ParseInt(pidstr, 10, 64)
	if !ok || err != nil {
		return errors.New("invalid pid")
	}
	pidstr = fmt.Sprint(pid)

	name := c.DefaultPostForm("name", "")
	if name == "" {
		return errors.New("bad name")
	}

	return o.mgr.Delete(pid, name)
}
