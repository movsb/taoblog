package post_translators

import (
	"bytes"
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
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// MarkdownTranslator ...
type MarkdownTranslator struct {
	PostTranslator
}

// Translate ...
func (me *MarkdownTranslator) Translate(source string, base string) (string, error) {
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
	buf := bytes.NewBuffer(nil)
	err := md.Convert([]byte(source), buf)
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
		panic(err)
	}
	width, height := imgConfig.Width, imgConfig.Height
	if strings.Contains(filepath.Base(path), `@2x.`) {
		width /= 2
		height /= 2
	}
	return width, height
}
