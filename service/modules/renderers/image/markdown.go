package image

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// 提供文章附件的引用来源
type AssetFinder func(path string) (name, url, description string, found bool)

type Image struct {
	assetFinder AssetFinder
}

func New(finder AssetFinder) *Image {
	return &Image{
		assetFinder: finder,
	}
}

func (e *Image) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(e, 100),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(e, 100),
	))
}

func (e *Image) RegisterFuncs(r renderer.NodeRendererFuncRegisterer) {
	r.Register(ast.KindImage, e.renderImage)
}

// Transform transforms the given AST tree.
func (e *Image) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	imagesToBeFigure := []ast.Node{}

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindImage:
			if n.Parent().ChildCount() == 1 {
				// 标记有来源的图片，移除其父 <p>。
				// 因为 <figure> 不能出现在 <p> 中。
				if e.assetFinder != nil {
					if url, err := url.Parse(string(n.(*ast.Image).Destination)); err == nil {
						if _, _, _, hasSource := e.assetFinder(url.Path); hasSource {
							imagesToBeFigure = append(imagesToBeFigure, n)
						}
					}
				}
			}
		}
		return ast.WalkContinue, nil
	})

	// 处理需要把 img 转换成 figure 的节点。
	for _, node := range imagesToBeFigure {
		p := node.Parent()
		pp := p.Parent()
		pp.ReplaceChild(pp, p, node)
	}
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

	// 如果有来源，包在 <figure> 中。
	//  <figure>
	//      <img src="full-piano.png" alt="Full Piano Keyboard">
	//      <figcaption>
	//          <a href="https://www.piano-keyboard-guide.com/piano-notes-and-keys.html" target="_blank" class="external">Full Piano Keyboard</a>
	//      </figcaption>
	//  </figure>
	//  defer 还能这么用！😂😂😂
	if e.assetFinder != nil {
		srcName, srcURL, srcDesc, hasSource := e.assetFinder(url.Path)
		if hasSource && srcName != "" && srcURL != "" {
			w.WriteString("<figure>\n")
			defer w.WriteString("</figure>\n")
			defer w.WriteString("</figcaption>\n")
			defer w.WriteString(fmt.Sprintf(
				`<a href="%s" target="_blank" class="external">%s</a>`,
				// TODO: srcURL | urlEscaper | attrEscaper
				util.EscapeHTML([]byte(srcURL)),
				util.EscapeHTML([]byte(srcName)),
			))
			defer w.WriteString("<figcaption>\n")
			_ = srcDesc
		}
	}

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
	_, _ = w.WriteString(` loading="lazy"`)

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

	_, _ = w.WriteString("/>")
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
