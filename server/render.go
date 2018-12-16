package main

import (
	"html/template"
	"io"
)

// Renderer is the blog theme renderer.
type Renderer struct {
	tmpl *template.Template
}

// NewRenderer ...
func NewRenderer(tmpl *template.Template) *Renderer {
	r := &Renderer{
		tmpl: tmpl,
	}
	return r
}

func (r *Renderer) mustExecuteTemplate(w io.Writer, name string, data interface{}) {
	err := r.tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		panic(err)
	}
}

// Render renders the named template with data data.
func (r *Renderer) Render(w io.Writer, name string, data interface{}) {
	r.mustExecuteTemplate(w, name, data)
}
