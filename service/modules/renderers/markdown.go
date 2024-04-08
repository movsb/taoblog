package renderers

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
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mathjax "github.com/litao91/goldmark-mathjax"
	wikitable "github.com/movsb/goldmark-wiki-table"
	"github.com/movsb/taoblog/modules/exception"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	html5 "golang.org/x/net/html"
)

// Markdown ...
type _Markdown struct {
	pathResolver       PathResolver
	removeTitleHeading bool // 是否移除 H1
	disableHeadings    bool // 评论中不允许标题
	disableHTML        bool // 禁止 HTML 元素
	openLinksInNewTab  bool // 新窗口打开链接
}

var (
	imageKind ast.NodeKind
)

func init() {
	imageKind = ast.NewNodeKind(`image`)
}

type Option func(me *_Markdown) error

// 解析 Markdown 中的相对链接。
func WithPathResolver(pathResolver PathResolver) Option {
	return func(me *_Markdown) error {
		me.pathResolver = pathResolver
		return nil
	}
}

// 移除 Markdown 中的标题（适用于文章）。
func WithRemoveTitleHeading(remove bool) Option {
	return func(me *_Markdown) error {
		me.removeTitleHeading = remove
		return nil
	}
}

// 不允许评论中存在任何级别的“标题”。
func WithDisableHeadings(disable bool) Option {
	return func(me *_Markdown) error {
		me.disableHeadings = disable
		return nil
	}
}

// 不允许使用 HTML 标签。
func WithDisableHTML(disable bool) Option {
	return func(me *_Markdown) error {
		me.disableHTML = disable
		return nil
	}
}

// 新窗口打开链接。
// TODO 目前只能针对 Markdown 链接， HTML 标签链接不可用。
func WithOpenLinksInNewTab() Option {
	return func(me *_Markdown) error {
		me.openLinksInNewTab = true
		return nil
	}
}

func NewMarkdown(options ...Option) *_Markdown {
	me := &_Markdown{}
	for _, option := range options {
		if err := option(me); err != nil {
			log.Println(err)
		}
	}
	return me
}

func (me *_Markdown) Render(source string) (string, string, error) {
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithExtensions(extension.DefinitionList),
		goldmark.WithExtensions(extension.Footnote),
		goldmark.WithExtensions(mathjax.MathJax),
		goldmark.WithExtensions(wikitable.New()),
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
	// TODO 移除这个循环，换 AstWalk
	for p := doc.FirstChild(); p != nil && n < maxDepth; n++ {
		switch {
		case p.Kind() == ast.KindHeading:
			heading := p.(*ast.Heading)
			switch heading.Level {
			case 1:
				title = string(heading.Text(sourceBytes))
				if !me.disableHeadings && me.removeTitleHeading {
					p = p.NextSibling()
					parent := heading.Parent()
					parent.RemoveChild(parent, heading)
					// p 已经 next，否则循环结束的时候再 next 会出错
					continue
				}
			}
		}
		p = p.NextSibling()
	}
	if n == maxDepth {
		panic(`max depth`)
	}

	if err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindImage:
				oldImage := n.(*ast.Image)
				newImage := &_Image{
					image: oldImage,
				}
				parent := n.Parent()
				parent.ReplaceChild(parent, oldImage, newImage)
			case ast.KindHeading:
				if me.disableHeadings {
					panic(exception.NewValidationError(`Markdown 不能包含标题`))
				}
			case ast.KindHTMLBlock, ast.KindRawHTML:
				if me.disableHTML {
					panic(exception.NewValidationError(`Markdown 不能包含 HTML 元素`))
				}
			case ast.KindAutoLink, ast.KindLink:
				if me.openLinksInNewTab {
					n.SetAttributeString(`target`, `_blank`)
					// TODO 会覆盖已经有了的
					n.SetAttributeString(`class`, `external`)
				}
			}
		}
		return ast.WalkContinue, nil
	}); err != nil {
		panic(err)
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
func (n *_Image) Type() ast.NodeType            { return ast.TypeInline }
func (n *_Image) Kind() ast.NodeKind            { return imageKind }

func (me *_Markdown) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*_Image)

	// 解析可能的自定义。
	// 不是很严格，可能有转义错误。
	url, _ := url.Parse(string(n.image.Destination))
	if url == nil {
		w.WriteString(`<img />`)
		return ast.WalkContinue, nil
	}

	styles := map[string]string{}
	classes := []string{}

	q := url.Query()
	scale := 1.0
	if q.Has(`float`) {
		styles[`float`] = `right`
		classes = append(classes, `f-r`)
		q.Del(`float`)
	}
	if n, err := strconv.ParseFloat(url.Query().Get(`scale`), 64); err == nil && n > 0 {
		scale = n
		q.Del(`scale`)
	}

	url.RawQuery = q.Encode()

	_, _ = w.WriteString("<img src=\"")
	// TODO 不知道 escape 几次了个啥。
	_, _ = w.Write(util.EscapeHTML(util.URLEscape([]byte(url.String()), true)))
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(nodeToHTMLText(n.image, source))
	_ = w.WriteByte('"')
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}
	_, _ = w.WriteString(` loading="lazy"`)

	path := url.Path
	if me.pathResolver != nil && !strings.Contains(path, `://`) {
		path2, err := me.pathResolver.Resolve(path)
		if err == nil {
			path = path2
		}
	}

	if len(styles) > 0 {
		b := strings.Builder{}
		b.WriteString(`style="`)
		for k, v := range styles {
			b.WriteString(fmt.Sprintf(`%s: %s;`, k, v))
		}
		b.WriteString(`"`)
		w.WriteString(b.String())
	}

	if len(classes) > 0 {
		w.WriteString(fmt.Sprintf(`class="%s"`, strings.Join(classes, " ")))
	}

	width, height := size(path)
	if width > 0 && height > 0 {
		w.WriteString(fmt.Sprintf(` width=%d height=%d`, int(float64(width)*scale), int(float64(height)*scale)))
	}

	// TODO 用 Node 序列化，代替手写标签。
	img := html5.Node{}
	_ = img

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

	for i := 1; i <= 10; i++ {
		scaleFmt := fmt.Sprintf(`@%dx.`, i)
		if strings.Contains(filepath.Base(path), scaleFmt) {
			width /= i
			height /= i
			break
		}
	}

	return width, height
}
