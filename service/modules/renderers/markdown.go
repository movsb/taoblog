package renderers

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	wikitable "github.com/movsb/goldmark-wiki-table"
	"github.com/movsb/taoblog/modules/utils"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	xnethtml "golang.org/x/net/html"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Markdown ...
type _Markdown struct {
	opts []Option2

	// 从内容中解析到的标题。
	// 外部初始化，导出。
	title *string

	removeTitleHeading bool // 是否移除 H1
	disableHeadings    bool // 评论中不允许标题
	disableHTML        bool // 禁止 HTML 元素

	openLinksInNewTab OpenLinksInNewTabKind // 新窗口打开链接

	modifiedAnchorReference string
	assetSourceFinder       AssetFinder

	noRendering bool
}

// TODO 不要返回 error。
// apply 的时候统一 catch 并返回初始化失败。
type Option func(me *_Markdown) error
type OptionNoError func(me *_Markdown)

// 移除 Markdown 中的标题（适用于文章）。
func WithRemoveTitleHeading() Option {
	return func(me *_Markdown) error {
		me.removeTitleHeading = true
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

// 不动态计算图片大小。适用于提交的时候，只会检查合法性。计算是在返回的时候进行。
// 不渲染，只解析，并判断合法性。不返回内容。
func WithoutRendering() Option {
	return func(me *_Markdown) error {
		me.noRendering = true
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
// 注意：锚点 （#section）这种始终不会在新窗口打开。
func WithOpenLinksInNewTab(kind OpenLinksInNewTabKind) Option {
	return func(me *_Markdown) error {
		me.openLinksInNewTab = kind
		return nil
	}
}

type OpenLinksInNewTabKind int

const (
	OpenLinksInNewTabKindKeep     OpenLinksInNewTabKind = iota // 不作为。
	OpenLinksInNewTabKindNever                                 // 全部链接在当前窗口打开。
	OpenLinksInNewTabKindAll                                   // 全部链接在新窗口打开，适用于评论预览时。
	OpenLinksInNewTabKindExternal                              // 仅外站链接在新窗口打开。
)

type AssetFinder func(path string) (name, url, description string, found bool)

// 提供文章附件的引用来源
func WithAssetSources(fn AssetFinder) Option {
	return func(me *_Markdown) error {
		me.assetSourceFinder = fn
		return nil
	}
}

func NewMarkdown(options ...any) *_Markdown {
	me := &_Markdown{}

	me.AddOptions(options...)

	return me
}

// TODO 判断重复。
func (me *_Markdown) AddOptions(options ...any) {
	for _, option := range options {
		if v1, ok := option.(Option); ok {
			if err := v1(me); err != nil {
				// TODO 处理错误。
				log.Println(err)
			}
		}
		if v1, ok := option.(OptionNoError); ok {
			v1(me)
		}
		me.opts = append(me.opts, option)
	}
}

// TODO 只是不渲染的话，其实不需要加载插件？
// TODO 把 parse、检查、渲染过程分开。
func (me *_Markdown) Render(source string) (string, error) {
	options := []goldmark.Option{
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			renderer.WithNodeRenderers(
				util.Prioritized(me, 100),
			),
		),
	}

	extensions := []goldmark.Extender{
		extension.GFM,
		extension.NewFootnote(extension.WithFootnoteBacklinkHTML(`^`)),
		wikitable.New(),
	}

	for _, opt := range me.opts {
		if tr, ok := opt.(goldmark.Extender); ok {
			extensions = append(extensions, tr)
		}
	}

	md := goldmark.New(append(options, goldmark.WithExtensions(extensions...))...)

	pCtx := parser.NewContext()

	for _, opt := range me.opts {
		if cp, ok := opt.(ContextPreparer); ok {
			cp.PrepareContext(pCtx)
		}
	}

	sourceBytes := []byte(source)
	doc := md.Parser().Parse(
		text.NewReader(sourceBytes),
		parser.WithContext(pCtx),
	)

	maxDepth := 10000 // this is to avoid unwanted infinite loop.
	n := 0
	// TODO 移除这个循环，换 AstWalk
	for p := doc.FirstChild(); p != nil && n < maxDepth; n++ {
		switch {
		case p.Kind() == ast.KindHeading:
			heading := p.(*ast.Heading)
			switch heading.Level {
			case 1:
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
				heading := n.(*ast.Heading)
				if me.title != nil && heading.Level == 1 {
					// 不允许重复定义标题
					if *me.title != "" {
						return ast.WalkStop, status.Errorf(codes.InvalidArgument, "内容中多次出现主标题")
					}
					*me.title = string(heading.Text(sourceBytes))
				}
				if me.disableHeadings {
					return ast.WalkStop, status.Errorf(codes.InvalidArgument, `Markdown 不能包含标题元素。`)
				}
			case ast.KindHTMLBlock, ast.KindRawHTML:
				if me.disableHTML {
					return ast.WalkStop, status.Errorf(codes.InvalidArgument, `Markdown 不能包含 HTML 标签。`)
				}
			case ast.KindAutoLink, ast.KindLink:
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
		return ``, err
	}

	// 处理需要把 img 转换成 figure 的节点。
	for _, node := range imagesToBeFigure {
		p := node.Parent()
		pp := p.Parent()
		pp.ReplaceChild(pp, p, node)
	}

	if me.openLinksInNewTab != OpenLinksInNewTabKindKeep {
		if err := me.doOpenLinkInNewTab(doc, []byte(source)); err != nil {
			return ``, err
		}
	}

	for _, opt := range me.opts {
		if walker, ok := opt.(EnteringWalker); ok {
			if err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering {
					return walker.WalkEntering(n)
				}
				return ast.WalkContinue, nil
			}); err != nil {
				panic(err)
			}
		}
	}

	if me.noRendering {
		return ``, nil
	}

	buf := bytes.NewBuffer(nil)
	err := md.Renderer().Render(buf, []byte(source), doc)
	if err != nil {
		return ``, err
	}

	htmlText := buf.Bytes()

	// 非常低效的接口。
	// TODO 重写一个新的 markdown 渲染器，渲染到 html 节点，而不是直接写 writer。
	for _, opt := range me.opts {
		if filter, ok := opt.(HtmlFilter); ok {
			htmlDoc, err := xnethtml.Parse(bytes.NewReader(htmlText))
			if err != nil {
				return ``, err
			}
			filtered, err := filter.FilterHtml(htmlDoc)
			if err != nil {
				return ``, err
			}
			htmlText = filtered
		}
	}

	if h2, err := gold_utils.ApplyHtmlTransformers(
		htmlText,
		utils.Map(
			utils.Filter(me.opts, func(o Option2) bool { return utils.Implements[gold_utils.HtmlTransformer](o) }),
			func(o Option2) gold_utils.HtmlTransformer { return o.(gold_utils.HtmlTransformer) },
		)...,
	); err != nil {
		return "", err
	} else {
		htmlText = h2
	}

	// TODO 和渲染分开，根本不是一个阶段的事
	prettified := ""
	for _, opt := range me.opts {
		if filter, ok := opt.(HtmlPrettifier); ok {
			if prettified != "" {
				return ``, errors.New(`不应有多个内容美化器`)
			}
			htmlDoc, err := xnethtml.Parse(bytes.NewReader(htmlText))
			if err != nil {
				return ``, err
			}
			filtered, err := filter.PrettifyHtml(htmlDoc)
			if err != nil {
				return ``, err
			}
			prettified = string(filtered)
		}
	}

	return utils.IIF(prettified == "", string(htmlText), prettified), err
}

// TODO 找到 body 之前的全部东西会被丢掉，比如注释，没啥问题
func renderHtmlDoc(doc *xnethtml.Node) ([]byte, error) {
	head, body := func() (head, body *xnethtml.Node) {
		// defer func() { recover() }()
		var walk func(node *xnethtml.Node)
		walk = func(node *xnethtml.Node) {
			switch node.Type {
			case xnethtml.ElementNode:
				if node.Data == `body` {
					body = node
					// panic("found body")
				} else if node.Data == `head` {
					head = node
				}
			}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
		walk(doc)
		return
	}()
	if body == nil && head == nil {
		return nil, errors.New(`empty html doc`)
	}
	buf := bytes.NewBuffer(nil)
	for c := head.FirstChild; c != nil; c = c.NextSibling {
		if err := xnethtml.Render(buf, c); err != nil {
			return nil, err
		}
	}
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := xnethtml.Render(buf, c); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (me *_Markdown) doOpenLinkInNewTab(doc ast.Node, source []byte) error {
	// Never 的时候只是简单地不处理。
	if me.openLinksInNewTab == OpenLinksInNewTabKindNever {
		return nil
	}

	addClass := func(node ast.Node) {
		var str string
		if cls, ok := node.AttributeString(`class`); ok {
			switch typed := cls.(type) {
			case string:
				str = typed
			case []byte:
				str = string(typed)
			}
		}
		if str == "" {
			str = `external`
		} else {
			str += ` external`
		}
		node.SetAttributeString(`class`, str)
		node.SetAttributeString(`target`, `_blank`)
	}

	modify := func(node ast.Node) {
		var dst string
		switch typed := node.(type) {
		case *ast.Link:
			dst = string(typed.Destination)
		case *ast.AutoLink:
			dst = string(typed.URL(source))
		}

		if me.openLinksInNewTab == OpenLinksInNewTabKindAll {
			if !strings.HasPrefix(dst, `#`) {
				addClass(node)
			}
			return
		} else if me.openLinksInNewTab == OpenLinksInNewTabKindExternal {
			// 外部站点新窗口打开。
			// 简单起见，默认站内都是相对链接。
			// 所以，如果不是相对，则总是外部的。
			if u, err := url.Parse(dst); err == nil {
				if u.Scheme != "" && u.Host != "" {
					addClass(node)
				}
			}
		}
	}

	return ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindAutoLink, ast.KindLink:
				modify(n)
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
	if me.assetSourceFinder != nil {
		srcName, srcURL, srcDesc, hasSource := me.assetSourceFinder(url.Path)
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
