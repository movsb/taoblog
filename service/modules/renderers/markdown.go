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
	urlpkg "net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	mathjax "github.com/litao91/goldmark-mathjax"
	wikitable "github.com/movsb/goldmark-wiki-table"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Markdown ...
type _Markdown struct {
	pathResolver       PathResolver
	removeTitleHeading bool // 是否移除 H1
	disableHeadings    bool // 评论中不允许标题
	disableHTML        bool // 禁止 HTML 元素
	openLinksInNewTab  bool // 新窗口打开链接

	modifiedAnchorReference string
	assetSourceFinder       AssetFinder

	useAbsolutePaths string
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

// 修改锚点页内引用（#）的指向为绝对地址。
// https://github.com/movsb/taoblog/blob/5c86466f3c1ab2f1543c3a5be4abc24f9c60c532/docs/TODO.md
func WithModifiedAnchorReference(relativePath string) Option {
	return func(me *_Markdown) error {
		me.modifiedAnchorReference = relativePath
		return nil
	}
}

// 新窗口打开链接。
// TODO 目前只能针对 Markdown 链接， HTML 标签链接不可用。
// 注意：锚点 （#section）这种不会在新窗口打开。
func WithOpenLinksInNewTab() Option {
	return func(me *_Markdown) error {
		me.openLinksInNewTab = true
		return nil
	}
}

type AssetFinder func(path string) (name, url, description string, found bool)

// 提供文章附件的引用来源
func WithAssetSources(fn AssetFinder) Option {
	return func(me *_Markdown) error {
		me.assetSourceFinder = fn
		return nil
	}
}

// 在 Tweets 页面下展示不止一篇文章的时候，文章内引用的资源的链接不能是相对链接（找不到），
// 必须修改成引用相对于文章的路径。
// 好希望有多个 <base> 支持啊，比如每个 <article> 下面有自己的 <base>。
// TODO：暂时只支持 <img>, <a>。
func WithUseAbsolutePaths(base string) Option {
	return func(me *_Markdown) error {
		me.useAbsolutePaths = base
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
			renderer.WithNodeRenderers(
				util.Prioritized(me, 100),
			),
		),
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithExtensions(extension.DefinitionList),
		goldmark.WithExtensions(extension.Footnote),
		goldmark.WithExtensions(mathjax.NewMathJax(
			mathjax.WithInlineDelim(`$`, `$`),
			mathjax.WithBlockDelim(`$$`, `$$`),
		)),
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

	imagesToBeFigure := []ast.Node{}

	if err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindHeading:
				if me.disableHeadings {
					return ast.WalkStop, status.Errorf(codes.InvalidArgument, `Markdown 不能包含标题元素。`)
				}
			case ast.KindHTMLBlock, ast.KindRawHTML:
				if me.disableHTML {
					return ast.WalkStop, status.Errorf(codes.InvalidArgument, `Markdown 不能包含 HTML 标签。`)
				}
			case ast.KindAutoLink, ast.KindLink:
				if n.Kind() == ast.KindLink {
					link := string(n.(*ast.Link).Destination)
					if !strings.HasPrefix(link, `#`) {
						if me.openLinksInNewTab {
							n.SetAttributeString(`target`, `_blank`)
							// TODO 会覆盖已经有了的
							n.SetAttributeString(`class`, `external`)
						}
					}
				}

				if n.Kind() == ast.KindLink && me.modifiedAnchorReference != "" {
					link := n.(*ast.Link)
					if href := string(link.Destination); strings.HasPrefix(href, "#") {
						if url, err := url.Parse(href); err == nil {
							url.Path = me.modifiedAnchorReference
							link.Destination = []byte(url.String())
						}
					}
				}
			case ast.KindImage:
				if n.Parent().ChildCount() == 1 {
					// 标记有来源的图片，移除其父 <p>。
					// 因为 <figure> 不能出现在 <p> 中。
					if me.assetSourceFinder != nil {
						if url, err := url.Parse(string(n.(*ast.Image).Destination)); err == nil {
							if _, _, _, hasSource := me.assetSourceFinder(url.Path); hasSource {
								imagesToBeFigure = append(imagesToBeFigure, n)
							}
						}
					}
				}
			}
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return ``, ``, err
	}

	// 处理需要把 img 转换成 figure 的节点。
	for _, node := range imagesToBeFigure {
		p := node.Parent()
		pp := p.Parent()
		pp.ReplaceChild(pp, p, node)
	}

	if me.useAbsolutePaths != "" {
		if err := me.doUseAbsolutePaths(doc); err != nil {
			return ``, ``, err
		}
	}

	buf := bytes.NewBuffer(nil)
	err := md.Renderer().Render(buf, []byte(source), doc)
	return title, buf.String(), err
}

func (me *_Markdown) doUseAbsolutePaths(doc ast.Node) error {
	base, _ := url.Parse(me.useAbsolutePaths)

	modify := func(u string) string {
		if u, err := url.Parse(u); err == nil {
			if u.Scheme == "" && u.Host == "" && !filepath.IsAbs(u.Path) {
				// 会丢失原 u 的查询参数。
				// return base.JoinPath(u.Path).String()
				u.Path = path.Join(base.Path, u.Path)
				return u.String()
			}
		}
		return u
	}

	return ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindImage:
				img := n.(*ast.Image)
				img.Destination = []byte(modify(string(img.Destination)))
			case ast.KindLink:
				link := n.(*ast.Link)
				link.Destination = []byte(modify(string(link.Destination)))
			}
		}
		return ast.WalkContinue, nil
	})
}

func (me *_Markdown) RegisterFuncs(r renderer.NodeRendererFuncRegisterer) {
	r.Register(ast.KindImage, me.renderImage)
}

func (me *_Markdown) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Image)

	// 解析可能的自定义。
	// 不是很严格，可能有转义错误。
	url, _ := url.Parse(string(n.Destination))
	if url == nil {
		w.WriteString(`<img />`)
		log.Println(`图片地址解析失败：`, string(n.Destination))
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

	// 如果有来源，包在 <figure> 中。
	//  <figure>
	//      <img src="full-piano.png" alt="Full Piano Keyboard">
	//      <figcaption>
	//          <a href="https://www.piano-keyboard-guide.com/piano-notes-and-keys.html" target="_blank" class="external">Full Piano Keyboard</a>
	//      </figcaption>
	//  </figure>
	//  defer 还能这么用！😂😂😂
	if me.assetSourceFinder != nil {
		srcName, srcURL, srcDesc, hasSource := me.assetSourceFinder(url.Path)
		if hasSource && srcName != "" && srcURL != "" {
			w.WriteString("<figure>\n")
			defer w.WriteString("</figure>\n")
			defer w.WriteString("</figcaption>\n")
			defer w.WriteString(fmt.Sprintf(
				`<a href="%s" target="_blank" class="external">%s</a>`,
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

	// 看起来是文章内的相对链接？
	// 如果是的话，需要 resolve 到相对应的目录。
	if url.Scheme == "" && url.Host == "" && me.pathResolver != nil {
		pathRelative := me.pathResolver.Resolve(url.Path)
		url.Path = pathRelative
	}

	width, height := size(url)
	if width > 0 && height > 0 {
		widthScaled, heightScaled := int(float64(width)*scale), int(float64(height)*scale)
		w.WriteString(fmt.Sprintf(` width=%d height=%d`, widthScaled, heightScaled))
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

func size(url *urlpkg.URL) (int, int) {
	var fp io.ReadCloser
	switch strings.ToLower(url.Scheme) {
	case `http`, `https`:
		resp, err := http.Get(url.String())
		if err != nil {
			panic(err)
		}
		fp = resp.Body
	default:
		// 有可能是引用别的文章的链接，这样会是以 / 开头的绝对路径。
		// 只是相对于站点，而不是相对于文件系统，所以要去除。
		f, err := os.Open(url.Path)
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
			if strings.EqualFold(filepath.Ext(url.Path), `.svg`) {
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

	// TODO 移除通过 @2x 类似的图片缩放支持。
	for i := 1; i <= 10; i++ {
		scaleFmt := fmt.Sprintf(`@%dx.`, i)
		if strings.Contains(filepath.Base(url.Path), scaleFmt) {
			width /= i
			height /= i
			break
		}
	}

	return width, height
}
