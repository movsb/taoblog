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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	pathResolver PathResolver
}

var (
	imageKind ast.NodeKind
)

func init() {
	imageKind = ast.NewNodeKind(`image`)
}

func (me *MarkdownTranslator) SetPathResolver(pathResolver PathResolver) {
	me.pathResolver = pathResolver
}

// Translate ...
func (me *MarkdownTranslator) Translate(source string) (string, string, error) {
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
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

	var title string
	maxDepth := 10000 // this is to avoid unwanted infinite loop.
	n := 0
	for p := doc.FirstChild(); p != nil && n < maxDepth; n++ {
		switch {
		case p.Kind() == ast.KindHeading:
			heading := p.(*ast.Heading)
			switch heading.Level {
			case 1:
				title = string(heading.Text(sourceBytes))
				p = p.NextSibling()
				parent := heading.Parent()
				parent.RemoveChild(parent, heading)
				continue
			}
		case p.Kind() == ast.KindParagraph:
			para := p.(*ast.Paragraph)
			for c := para.FirstChild(); c != nil; c = c.NextSibling() {
				if c.Kind() == ast.KindImage {
					oldImage := c.(*ast.Image)
					newImage := &_Image{
						image: oldImage,
					}
					para.ReplaceChild(para, oldImage, newImage)
					c = newImage
				}
			}
		}
		p = p.NextSibling()
	}
	if n == maxDepth {
		panic(`max depth`)
	}

	rdr := md.Renderer()
	if reg, ok := rdr.(renderer.NodeRendererFuncRegisterer); ok {
		reg.Register(imageKind, me.renderImage)
	}

	buf := bytes.NewBuffer(nil)
	err := rdr.Render(buf, []byte(source), doc)
	return title, buf.String(), err
}

type _Image struct {
	ast.BaseBlock
	image *ast.Image
}

func (n *_Image) Dump(source []byte, level int) { ast.DumpHelper(n, source, level, nil, nil) }
func (n *_Image) Type() ast.NodeType            { return ast.TypeBlock }
func (n *_Image) Kind() ast.NodeKind            { return imageKind }

func (me *MarkdownTranslator) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*_Image)
	_, _ = w.WriteString("<img src=\"")
	_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.image.Destination, true)))
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(nodeToHTMLText(n.image, source))
	_ = w.WriteByte('"')
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}
	_, _ = w.WriteString(` loading="lazy"`)

	path := string(n.image.Destination)
	if me.pathResolver != nil && !strings.Contains(path, `://`) {
		path2, err := me.pathResolver.Resolve(path)
		if err == nil {
			path = path2
		}
	}
	width, height := size(path)
	if width > 0 && height > 0 {
		w.WriteString(fmt.Sprintf(` width=%d height=%d`, width, height))
	}

	_, _ = w.WriteString(" />")
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

func size(path string) (int, int) {
	var fp io.ReadCloser
	if strings.Contains(path, `://`) {
		resp, err := http.Get(path)
		if err != nil {
			panic(err)
		}
		fp = resp.Body
	} else {
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
		log.Println(err)
		return 0, 0
	}
	width, height := imgConfig.Width, imgConfig.Height
	if strings.Contains(filepath.Base(path), `@2x.`) {
		width /= 2
		height /= 2
	}
	return width, height
}
