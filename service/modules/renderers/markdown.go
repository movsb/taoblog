package renderers

import (
	"bytes"
	_ "embed"
	"log"
	"net/url"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	xnethtml "golang.org/x/net/html"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Renderer interface {
	Render(source string) (string, error)
}

type HTML struct {
	Renderer
}

func (me *HTML) Render(source string) (string, error) {
	return source, nil
}

// Markdown ...
type _Markdown struct {
	opts []Option2

	// 从内容中解析到的标题。
	// 外部初始化，导出。
	title *string

	removeTitleHeading bool // 是否移除 H1
	disableHeadings    bool // 评论中不允许标题
	disableHTML        bool // 禁止 HTML 元素

	modifiedAnchorReference string

	noRendering bool
	noTransform bool
	xhtml       bool

	htmlPrettifier          HtmlPrettifier
	fencedCodeBlockRenderer map[string]gold_utils.FencedCodeBlockRenderer
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

func WithoutTransform() Option {
	return func(me *_Markdown) error {
		me.noTransform = true
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

func WithXHTML() Option {
	return func(me *_Markdown) error {
		me.xhtml = true
		return nil
	}
}

func NewMarkdown(options ...any) *_Markdown {
	me := &_Markdown{
		fencedCodeBlockRenderer: map[string]gold_utils.FencedCodeBlockRenderer{},
	}

	me.AddOptions(options...)

	// 总是添加辅助功能扩展。
	me.AddOptions(&gold_utils.FencedCodeBlockExtender{
		Renders: &me.fencedCodeBlockRenderer,
	})

	return me
}

// TODO 判断重复。
//
// TODO 添加具体的、有类型的函数。
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

func (me *_Markdown) AddHtmlTransformers(trs ...gold_utils.HtmlTransformer) {
	for _, tr := range trs {
		me.opts = append(me.opts, tr)
	}
}

// TODO 只是不渲染的话，其实不需要加载插件？
// TODO 把 parse、检查、渲染过程分开。
func (me *_Markdown) Render(source string) (_ string, outErr error) {
	defer utils.CatchAsError(&outErr)

	options := []goldmark.Option{
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	}

	if me.xhtml {
		options = append(options, goldmark.WithRendererOptions(html.WithXHTML()))
	}

	extensions := []goldmark.Extender{}

	for _, opt := range me.opts {
		if tr, ok := opt.(goldmark.Extender); ok {
			extensions = append(extensions, tr)
		}
	}

	md := goldmark.New(append(options, goldmark.WithExtensions(extensions...))...)

	sourceBytes := []byte(source)
	doc := md.Parser().Parse(text.NewReader(sourceBytes))

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

	utils.Must(ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
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
			}
		}
		return ast.WalkContinue, nil
	}))

	for _, opt := range me.opts {
		if walker, ok := opt.(EnteringWalker); ok {
			utils.Must(ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if entering {
					return walker.WalkEntering(n)
				}
				return ast.WalkContinue, nil
			}))
		}
	}

	if me.noRendering {
		return ``, nil
	}

	buf := bytes.NewBuffer(nil)
	utils.Must(md.Renderer().Render(buf, []byte(source), doc))

	htmlText := buf.Bytes()

	if !me.noTransform {
		htmlText = utils.Must1(gold_utils.ApplyHtmlTransformers(
			htmlText,
			utils.Map(
				utils.Filter(me.opts, func(o Option2) bool { return utils.Implements[gold_utils.HtmlTransformer](o) }),
				func(o Option2) gold_utils.HtmlTransformer { return o.(gold_utils.HtmlTransformer) },
			)...,
		))
	}

	if me.htmlPrettifier != nil {
		htmlText = utils.Must1(me.prettifyHtml(htmlText))
	}

	return string(htmlText), nil
}

func (me *_Markdown) prettifyHtml(raw []byte) (_ []byte, outErr error) {
	defer utils.CatchAsError(&outErr)
	htmlDoc := utils.Must1(xnethtml.Parse(bytes.NewReader(raw)))
	filtered := utils.Must1(me.htmlPrettifier.PrettifyHtml(htmlDoc))
	return filtered, nil
}
