package post_translators

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif" // shut up
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// MarkdownTranslator ...
type MarkdownTranslator struct {
	once sync.Once
}

var referencesKind ast.NodeKind

// Translate ...
func (me *MarkdownTranslator) Translate(cb *Callback, source string, base string) (string, error) {
	me.once.Do(func() {
		referencesKind = ast.NewNodeKind(`references`)
	})

	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithImageDataSrc(),
			html.WithImageSizeFunc(func(path string) (int, int) {
				return size(base, path)
			}),
		),
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithExtensions(extension.DefinitionList),
		goldmark.WithExtensions(extension.Footnote),
		goldmark.WithExtensions(mathjax.MathJax),
	)

	pCtx := parser.NewContext()
	sourceBytes := []byte(source)
	doc := md.Parser().Parse(
		text.NewReader(sourceBytes),
		parser.WithContext(pCtx),
	)

	if cb == nil {
		cb = &Callback{}
	}

	maxDepth := 10000 // this is to avoid unwanted infinite loop.
	n := 0
	for p := doc.FirstChild(); p != nil && n < maxDepth; n++ {
		switch {
		case p.Kind() == ast.KindHeading:
			heading := p.(*ast.Heading)
			switch heading.Level {
			case 1:
				title := string(heading.Text(sourceBytes))
				if cb.SetTitle != nil {
					cb.SetTitle(title)
				}

				p = p.NextSibling()
				parent := heading.Parent()
				parent.RemoveChild(parent, heading)
				continue
			}
		case p.Kind() == ast.KindParagraph:
			para := p.(*ast.Paragraph)
			if para.Lines().Len() != 1 {
				break
			}
			line := para.Lines().At(0)
			raw := string(sourceBytes[line.Start:line.Stop])
			if raw == `[REFERENCES]` {
				n := &_Ref{
					raw: genRefs(pCtx.References()),
				}
				p.Parent().ReplaceChild(p.Parent(), p, n)
				p = n
				continue
			}
		}
		p = p.NextSibling()
	}
	if n == maxDepth {
		panic(`max depth`)
	}

	rdr := md.Renderer()
	if reg, ok := rdr.(renderer.NodeRendererFuncRegisterer); ok {
		reg.Register(referencesKind, renderRefs)
	}

	buf := bytes.NewBuffer(nil)
	err := rdr.Render(buf, []byte(source), doc)
	return buf.String(), err
}

func genRefs(refs []parser.Reference) string {
	t := template.Must(template.New(`gen-refs`).Parse(`<li><a title="{{printf "%s" .Title}}" href="{{printf "%s" .Destination}}">{{or .Title .Destination | printf "%s"}}</a></li>`))
	w := bytes.NewBufferString("<ol class=references>\n")
	for _, ref := range refs {
		err := t.Execute(w, ref)
		if err != nil {
			panic(err)
		}
		w.WriteByte('\n')
	}
	w.WriteString(`</ol>`)
	return w.String()
}

type _Ref struct {
	ast.BaseBlock
	raw string
}

// Dump implements Node.Dump .
func (n *_Ref) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Type implements Node.Type .
func (n *_Ref) Type() ast.NodeType {
	return ast.TypeBlock
}

// Kind implements Node.Kind.
func (n *_Ref) Kind() ast.NodeKind {
	return referencesKind
}

func renderRefs(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	r := n.(*_Ref)
	_, err := writer.WriteString(r.raw)
	return ast.WalkContinue, err
}

func size(base string, path string) (int, int) {
	var fp io.ReadCloser
	if strings.Contains(path, `:`) {
		resp, err := http.Get(path)
		if err != nil {
			panic(err)
		}
		fp = resp.Body
	} else {
		path = filepath.Join(base, path)
		f, err := os.Open(path)
		if err != nil {
			// panic(err)
			return 0, 0
		}
		fp = f
	}
	defer fp.Close()
	imgConfig, _, err := image.DecodeConfig(fp)
	if err != nil {
		if sfp, ok := fp.(io.ReadSeeker); ok {
			if _, err := sfp.Seek(0, io.SeekStart); err != nil {
				panic(err)
			}
			if strings.EqualFold(filepath.Ext(path), `.svg`) {
				type _SvgSize struct {
					Width  string `xml:"width,attr"`
					Height string `xml:"height,attr"`
				}
				ss := _SvgSize{}
				if err := xml.NewDecoder(sfp).Decode(&ss); err != nil {
					panic(err)
				}
				var w, h int
				fmt.Sscanf(ss.Width, `%d`, &w)
				fmt.Sscanf(ss.Height, `%d`, &h)
				return w, h
			}
		}
		panic(err)
	}
	width, height := imgConfig.Width, imgConfig.Height
	if strings.Contains(filepath.Base(path), `@2x.`) {
		width /= 2
		height /= 2
	}
	return width, height
}
