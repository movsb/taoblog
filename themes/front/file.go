package front

import (
	"os"

	"github.com/gin-gonic/gin"
)

var (
	fileHost = os.Getenv("FILE_HOST")
)

func (f *Front) queryByFile(c *gin.Context, postID int64, file string) {
	user := f.auth.AuthCookie(c)
	path := f.service.GetFile(postID, file)
	if !user.IsGuest() {
		if _, err := os.Stat(path); err == nil {
			c.File(path)
			return
		}
	}
	remotePath := fileHost + "/" + path
	c.Redirect(302, remotePath)
}
