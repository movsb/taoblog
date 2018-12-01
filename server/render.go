package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

type IRendererData interface {
	PageType() string
}

type Renderer struct {
}

func NewRenderer() *Renderer {
	r := &Renderer{}
	return r
}

func (r *Renderer) execTemplate(c *gin.Context, t string, d IRendererData) {
	log.Println("before executing post template")
	if err := templates.ExecuteTemplate(c.Writer, t, d); err != nil {
		panic(err)
	}
	log.Println("after executing post template")
}

func (r *Renderer) RenderHome(c *gin.Context, home *Home) {
	r.execTemplate(c, "home", home)
}

func (r *Renderer) RenderPost(c *gin.Context, post *Post) {
	r.execTemplate(c, "single", post)
}

func (r *Renderer) RenderTags(c *gin.Context, posts *QueryTags) {
	r.execTemplate(c, "tags", posts)
}
