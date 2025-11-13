package renderers

import (
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/ast"
	"golang.org/x/net/html"

	_ "github.com/movsb/taoblog/service/modules/renderers/anniversary"
)

// 后面统一改成 Option。
type Option2 = any

type EnteringWalker interface {
	WalkEntering(n ast.Node) (ast.WalkStatus, error)
}

type HtmlPrettifier interface {
	PrettifyHtml(doc *html.Node) ([]byte, error)
}

// -----------------------------------------------------------------------------

// 获取从 Markdown 中解析得到的一级标题。
func WithTitle(title *string) OptionNoError {
	return func(me *_Markdown) {
		me.title = title
	}
}

// -----------------------------------------------------------------------------

func WithFencedCodeBlockRenderer(language string, r gold_utils.FencedCodeBlockRenderer) OptionNoError {
	return func(me *_Markdown) {
		me.fencedCodeBlockRenderer[language] = r
	}
}

func WithHtmlPrettifier(p HtmlPrettifier) OptionNoError {
	return func(me *_Markdown) {
		me.htmlPrettifier = p
	}
}

func WithoutCJK() OptionNoError {
	return func(me *_Markdown) {
		me.withoutCJK = true
	}
}
