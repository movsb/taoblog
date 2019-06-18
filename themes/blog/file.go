package blog

import (
	"os"

	"github.com/gin-gonic/gin"
)

var (
	fileHost = os.Getenv("FILE_HOST")
)

func (b *Blog) queryByFile(c *gin.Context, postID int64, file string) {
	user := b.auth.AuthCookie(c)
	path := b.service.GetFile(postID, file)
	if user.IsAdmin() || fileHost == "" {
		c.File(path)
	} else {
		remotePath := fileHost + "/" + path
		c.Redirect(302, remotePath)
	}
}
