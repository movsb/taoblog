package image

import (
	"bytes"
	"embed"
	"fmt"
	std_html "html"
	"io"
	"log"
	"mime"
	"net/url"
	"path"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css script.js
var _embed embed.FS
var _local = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `image`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
		dynamic.WithScripts(module, `script.js`)
	})
}

type Image struct {
	web gold_utils.WebFileSystem
}

func New(web gold_utils.WebFileSystem) *Image {
	return &Image{
		web: web,
	}
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

	fileType := mime.TypeByExtension(path.Ext(url.Path))
	switch {
	case strings.HasPrefix(fileType, `video/`):
		renderVideo(w, url)
	case strings.HasPrefix(fileType, `audio/`):
		renderAudio(w, url)
	default:
		if strings.HasSuffix(url.Path, `.table`) {
			e.renderTable(w, url)
		} else {
			renderImage(w, url, n, source)
		}
	}

	return ast.WalkSkipChildren, nil
}

func (e *Image) renderTable(w util.BufWriter, url *url.URL) {
	fp, err := e.web.OpenURL(url.String())
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()
	io.Copy(w, fp)
}

func renderVideo(w util.BufWriter, url *url.URL) {
	fmt.Fprintf(w, `<video controls src="%s"></video>`, std_html.EscapeString(url.String()))
}

func renderAudio(w util.BufWriter, url *url.URL) {
	fmt.Fprintf(w, `<audio controls src="%s"></audio>`, std_html.EscapeString(url.String()))
}

func renderImage(w util.BufWriter, url *url.URL, n *ast.Image, source []byte) {
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
	// 模糊效果/闪图。
	// 注意：不要给 live-photo 加，会自动去掉。
	if q.Has(`blur`) {
		classes = append(classes, `blur`)
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
