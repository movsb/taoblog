package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type FileUpload struct {
	folder string
}

func NewFileUpload(folder string) *FileUpload {
	return &FileUpload{
		folder: folder,
	}
}

func (o *FileUpload) Upload(c *gin.Context) error {
	pidstr, ok := c.GetPostForm("pid")
	pid, err := strconv.Atoi(pidstr)
	if !ok || err != nil {
		return errors.New("invalid pid")
	}
	pidstr = fmt.Sprint(pid)

	root := filepath.Join(o.folder, pidstr)
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	for _, file := range form.File["files[]"] {
		path := filepath.Join(root, file.Filename)
		if err = c.SaveUploadedFile(file, path); err != nil {
			return err
		}
	}

	return nil
}

func (o *FileUpload) List(c *gin.Context) []string {
	var files = make([]string, 0)

	pidstr, ok := c.GetQuery("pid")
	pid, err := strconv.Atoi(pidstr)
	if !ok || err != nil {
		return files
	}
	pidstr = fmt.Sprint(pid)

	root := filepath.Join(o.folder, pidstr)
	err = filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			files = append(files, info.Name())
		}

		return nil
	})

	return files
}

func (o *FileUpload) Delete(c *gin.Context) bool {
	pidstr, ok := c.GetPostForm("pid")
	pid, err := strconv.Atoi(pidstr)
	if !ok || err != nil {
		return false
	}
	pidstr = fmt.Sprint(pid)

	name := c.DefaultPostForm("name", "")
	if name == "" {
		return false
	}

	root := filepath.Join(o.folder, pidstr, name)
	err = os.Remove(root)
	return err == nil
}
