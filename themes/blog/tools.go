package blog

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func toolpath(name string) string {
	// TODO use base
	return filepath.Join("blog/tools/bin", name)
}

func aes2htm(c *gin.Context) {
	source := c.DefaultPostForm("source", "")
	path := toolpath("aes2htm")
	cmd := exec.Command(path)
	strread := strings.NewReader(source)
	cmd.Stdin = strread
	outBytes, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	c.String(200, "%s", string(outBytes))
}
