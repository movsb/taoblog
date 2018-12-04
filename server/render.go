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

func (r *Renderer) Render(c *gin.Context, name string, d IRendererData) {
	r.execTemplate(c, name, d)
}
