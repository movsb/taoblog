package renderers

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark/ast"
	"golang.org/x/net/html"
)

// 后面统一改成 Option。
type Option2 = any

type EnteringWalker interface {
	WalkEntering(n ast.Node) (ast.WalkStatus, error)
}

type HtmlFilter interface {
	FilterHtml(htmlText []byte) ([]byte, error)
}

// -----------------------------------------------------------------------------

func appendClass(node ast.Node, name string) {
	any, found := node.AttributeString(name)
	if list, ok := any.(string); !found || ok {
		if list == "" {
			list = name
		} else {
			list += " "
			list += name
		}
		node.SetAttributeString(`class`, list)
	}
}

func Testing() Option {
	return func(me *_Markdown) error {
		me.testing = true
		return nil
	}
}

// -----------------------------------------------------------------------------

type _ReserveListItemMarkerStyle struct{}

var knownListItemMarkers = map[byte]string{
	'-': `minus`,
	'+': `plus`,
	'*': `asterisk`,
	'.': `period`,
	')': `parenthesis`,
}

func (*_ReserveListItemMarkerStyle) WalkEntering(n ast.Node) (ast.WalkStatus, error) {
	switch typed := n.(type) {
	case *ast.List:
		if class, ok := knownListItemMarkers[typed.Marker]; ok {
			appendClass(typed, `marker-`+class)
		}
	}
	return ast.WalkContinue, nil
}

// 保留列表样式。
//
// 只是增加类名，前端通过类名自行决定怎么展示。
func WithReserveListItemMarkerStyle() Option2 {
	return &_ReserveListItemMarkerStyle{}
}

// -----------------------------------------------------------------------------

var prettifierStrings = map[string]string{
	`a`:      `链接`,
	`img`:    `图片`,
	`table`:  `表格`,
	`iframe`: `页面`,
	`video`:  `视频`,
	`audio`:  `音频`,
	`canvas`: `画布`,
	`embed`:  `对象`,
	`map`:    `地图`,
	`object`: `对象`,
	`script`: `脚本`,
	`svg`:    `图片`,
}

var prettifierFuncs = map[string]func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus{
	`div`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		for _, a := range node.Attr {
			if strings.ToLower(a.Key) == `class` {
				if strings.Contains(a.Val, `footnotes`) {
					return ast.WalkSkipChildren
				}
			}
		}
		return ast.WalkContinue
	},
	`pre`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		if node.FirstChild != nil && node.FirstChild.NextSibling == nil && node.FirstChild.Data == `code` {
			buf.WriteString(`[代码]`)
			return ast.WalkSkipChildren
		}
		return ast.WalkContinue
	},
}

type _ContentPrettifier struct{}

func (*_ContentPrettifier) FilterHtml(htmlText []byte) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(htmlText))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)

	var walk func(buf *bytes.Buffer, node *html.Node)
	walk = func(buf *bytes.Buffer, node *html.Node) {
		walkStatus := ast.WalkContinue
		switch node.Type {
		case html.TextNode:
			buf.WriteString(node.Data)
		case html.ElementNode:
			walkStatus = ast.WalkContinue
			if f, ok := prettifierStrings[node.Data]; ok {
				buf.WriteString(fmt.Sprintf(`[%s]`, f))
				walkStatus = ast.WalkSkipChildren
			} else if f, ok := prettifierFuncs[node.Data]; ok {
				walkStatus = f(buf, node)
			}
		}
		if walkStatus == ast.WalkContinue {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				walk(buf, c)
			}
		}
	}

	walk(buf, doc)

	return buf.Bytes(), nil
}

// 简化并美化 Markdown 的展示。
// 比如：
// - 不显示复杂的链接、图片、表格、代码块等元素。
// - 不显示脚注。
//
// NOTE 因为 Markdown 只能解析而不能还原，所以这里处理的是 HTML 内容。
// HTML 可以在 NODE 之间等价转换。
// 而由于 Markdown 目前在转换成 HTML 时采用了后端渲染代码。
// 所以 chroma 把 code 包裹在了 table 中。需要特别处理。
//
// https://yari-demos.prod.mdn.mozit.cloud/en-US/docs/Web/HTML/Inline_elements
func WithHtmlPrettifier() Option2 {
	return &_ContentPrettifier{}
}