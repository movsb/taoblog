package toc

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/toc"
)

func New(toc *[]byte) goldmark.Extender {
	return &ToC{
		out: toc,
	}
}

type ToC struct {
	md  goldmark.Markdown
	toc *toc.TOC
	out *[]byte
}

func (e *ToC) Extend(m goldmark.Markdown) {
	e.md = m
	m.Parser().AddOptions(
		parser.WithAutoHeadingID(),
		parser.WithASTTransformers(util.Prioritized(e, 1000)),
	)
}

func (e *ToC) Transform(doc *ast.Document, reader text.Reader, ctx parser.Context) {
	t, err := toc.Inspect(doc, reader.Source(), toc.MinDepth(2), toc.Compact(true))
	if err != nil {
		// There are currently no scenarios under which Inspect
		// returns an error but we have to account for it anyway.
		return
	}

	e.toc = t

	if e.out != nil {
		listNode := toc.RenderList(t)
		buf := bytes.NewBuffer(nil)
		e.md.Renderer().Render(buf, nil, listNode)
		*e.out = buf.Bytes()
	}
}
