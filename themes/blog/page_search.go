package blog

import (
	"github.com/gin-gonic/gin"
)

func (b *Blog) getPageSearch(c *gin.Context) {
	b.render(c.Writer, `search`, b.cfg.Site.Search)
}
