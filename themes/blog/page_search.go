package blog

import (
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (b *Blog) getPageSearch(c *gin.Context) {
	c.File(filepath.Join(b.base, "statics/search.html"))
}
