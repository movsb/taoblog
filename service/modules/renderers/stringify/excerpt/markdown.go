package excerpt

import (
	"bytes"

	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type ExcerptGenerator struct {
	excerpt *string
}

func New(output *string) *ExcerptGenerator {
	return &ExcerptGenerator{
		excerpt: output,
	}
}

func (m *ExcerptGenerator) WalkEntering(n ast.Node, source []byte) (ast.WalkStatus, error) {
	if n.Kind() != ast.KindParagraph {
		return ast.WalkContinue, nil
	}

	p := n.(*ast.Paragraph)

	hr := html.NewRenderer(html.WithUnsafe(), html.WithEastAsianLineBreaks(html.EastAsianLineBreaksSimple))
	r := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(hr, 1)))

	buf := bytes.NewBuffer(nil)
	r.Render(buf, source, p)

	*m.excerpt = m.textOf(buf.Bytes())

	// 只取第一个段落。
	return ast.WalkStop, nil
}

func (m *ExcerptGenerator) textOf(p []byte) string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(p))
	if err != nil {
		return ``
	}
	return doc.Text()
}
