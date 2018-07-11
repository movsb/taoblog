package main

import (
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
		EndReq(c, false, "expect: source")
		return
	}

	path := toolpath("aes2htm")
	cmd := exec.Command(path)
	strread := strings.NewReader(source)
	cmd.Stdin = strread
	outBytes, err := cmd.Output()
	EndReq(c, err, string(outBytes))
}
