package main

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func toolpath(name string) string {
	return filepath.Join(config.base, "tools/bin", name)
}

func aes2htm(c *gin.Context) {
	source, ok := c.GetPostForm("source")
	if !ok {
		finishError(c, -1, errors.New("expect: source"))
		return
	}

	path := toolpath("aes2htm")
	cmd := exec.Command(path)
	strread := strings.NewReader(source)
	cmd.Stdin = strread
	outBytes, err := cmd.Output()
	if err != nil {
		finishError(c, -1, err)
		return
	}
	finishDone(c, 0, "", string(outBytes))
}
