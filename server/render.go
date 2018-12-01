package main

import (
	"html/template"

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
	funcs := template.FuncMap{
		"get_config": func(name string) string {
			return optmgr.GetDef(gdb, name, "")
		},
	}
	templates, err := template.New("taoblog").Funcs(funcs).ParseGlob("../theme/*.html")
	if err != nil {
		panic(err)
	}
	if err := templates.ExecuteTemplate(c.Writer, t, d); err != nil {
		panic(err)
	}
}

func (r *Renderer) RenderHome(c *gin.Context, home *Home) {
	r.execTemplate(c, "home", home)
}

func (r *Renderer) RenderPost(c *gin.Context, post *Post) {
	r.execTemplate(c, "single", post)
}
