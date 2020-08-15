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

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// MarkdownTranslator ...
type MarkdownTranslator struct {
}

// Translate ...
func (me *MarkdownTranslator) Translate(cb *Callback, source string, base string) (string, error) {
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

	sourceBytes := []byte(source)
	doc := md.Parser().Parse(text.NewReader(sourceBytes))

	if cb == nil {
		cb = &Callback{}
	}

	if cb.SetTitle != nil {
		for p := doc.FirstChild(); p != nil; p = p.NextSibling() {
			if p.Kind() == ast.KindHeading {
				heading := p.(*ast.Heading)
				switch heading.Level {
				case 1:
					title := string(heading.Text(sourceBytes))
					cb.SetTitle(title)

					parent := heading.Parent()
					parent.RemoveChild(parent, heading)
				}
			}
		}
	}

	buf := bytes.NewBuffer(nil)
	err := md.Renderer().Render(buf, []byte(source), doc)
	return buf.String(), err
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
