package blog

import (
	"os"

	"github.com/gin-gonic/gin"
)

func (b *Blog) queryByFile(c *gin.Context, postID int64, file string) {
	path := b.service.GetFile(postID, file)

	redir := true

	fileHost := b.cfg.Data.File.Mirror

	// remote isn't enabled, use local only
	if redir && fileHost == "" {
		redir = false
	}
	// when logged in, see the newest-uploaded file
	if redir && b.auth.AuthCookie(c).IsAdmin() {
		redir = false
	}
	// if no referer, don't let them know we're using file host
	if redir && c.GetHeader("Referer") == "" {
		redir = false
	}
	// if file isn't in local, we should redirect
	if !redir {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			redir = true
		}
	}

	if redir {
		remotePath := fileHost + "/" + path
		c.Redirect(307, remotePath)
	} else {
		c.File(path)
	}
}
