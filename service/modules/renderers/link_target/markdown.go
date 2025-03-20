package link_target

import (
	"net/url"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type OpenLinksInNewTabKind int

const (
	OpenLinksInNewTabKindKeep     OpenLinksInNewTabKind = iota // 不作为。
	OpenLinksInNewTabKindAll                                   // 全部链接在新窗口打开，适用于评论预览时。
	OpenLinksInNewTabKindExternal                              // 仅外站链接在新窗口打开。
)

// 新窗口打开链接。
// TODO 目前只能针对 Markdown 链接， HTML 标签链接不可用。
// 注意：锚点 （#section）这种始终不会在新窗口打开。
type Extender struct {
	openLinksInNewTab OpenLinksInNewTabKind // 新窗口打开链接
}

func New(kind OpenLinksInNewTabKind) *Extender {
	return &Extender{
		openLinksInNewTab: kind,
	}
}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(e, 100),
	))
}

// Transform transforms the given AST tree.
func (e *Extender) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	// Never 的时候只是简单地不处理。
	if e.openLinksInNewTab == OpenLinksInNewTabKindKeep {
		return
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
			dst = string(typed.URL(reader.Source()))
		}

		if e.openLinksInNewTab == OpenLinksInNewTabKindAll {
			if !strings.HasPrefix(dst, `#`) {
				addClass(node)
			}
			return
		} else if e.openLinksInNewTab == OpenLinksInNewTabKindExternal {
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

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindAutoLink, ast.KindLink:
				modify(n)
			}
		}
		return ast.WalkContinue, nil
	})
}
