package front

import (
	"github.com/gin-gonic/gin"
)

func (f *Front) getPageSearch(c *gin.Context) {
	c.File("front/statics/search.html")
}
