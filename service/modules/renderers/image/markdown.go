package image

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type Image struct {
}

func New() *Image {
	return &Image{}
}

func (e *Image) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(e, 100),
	))
}

func (e *Image) RegisterFuncs(r renderer.NodeRendererFuncRegisterer) {
	r.Register(ast.KindImage, e.renderImage)
}

func (e *Image) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Image)

	// 解析可能的自定义。
	// 不是很严格，可能有转义错误。
	url, _ := url.Parse(string(n.Destination))
	if url == nil {
		w.WriteString(`<img>`)
		log.Println(`图片地址解析失败：`, string(n.Destination))
		return ast.WalkContinue, nil
	}

	styles := map[string]string{}
	classes := []string{}

	q := url.Query()
	if q.Has(`float`) {
		styles[`float`] = `right`
		classes = append(classes, `f-r`)
		q.Del(`float`)
	}
	if q.Has(`t`) {
		classes = append(classes, `transparent`)
		q.Del(`t`)
	}

	url.RawQuery = q.Encode()

	_, _ = w.WriteString("<img src=\"")
	// TODO 不知道 escape 几次了个啥。
	_, _ = w.Write(util.EscapeHTML(util.URLEscape([]byte(url.String()), true)))
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(nodeToHTMLText(n, source))
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		w.Write(util.EscapeHTML(n.Title))
		_ = w.WriteByte('"')
	}
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}

	if len(styles) > 0 {
		b := strings.Builder{}
		b.WriteString(` style="`)
		for k, v := range styles {
			b.WriteString(fmt.Sprintf(`%s: %s;`, k, v))
		}
		b.WriteString(`"`)
		w.WriteString(b.String())
	}

	if len(classes) > 0 {
		w.WriteString(fmt.Sprintf(` class="%s"`, strings.Join(classes, " ")))
	}

	_, _ = w.WriteString(">")
	return ast.WalkSkipChildren, nil
}

func nodeToHTMLText(n ast.Node, source []byte) []byte {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok && s.IsCode() {
			buf.Write(s.Text(source))
		} else if !c.HasChildren() {
			buf.Write(util.EscapeHTML(c.Text(source)))
		} else {
			buf.Write(nodeToHTMLText(c, source))
		}
	}
	return buf.Bytes()
}
