package stringify

import (
	"bytes"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark/ast"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var prettifierStrings = map[string]string{
	`img`:    `图片`,
	`table`:  `表格`,
	`video`:  `视频`,
	`audio`:  `音频`,
	`canvas`: `画布`,
	`embed`:  `对象`,
	`map`:    `地图`,
	`object`: `对象`,
	`script`: `脚本`,
	`svg`:    `图片`,
	`code`:   `代码`,
}

var prettifierFuncs = map[string]func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus{
	`a`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		hrefIndex := slices.IndexFunc(node.Attr, func(attr html.Attribute) bool { return attr.Key == `href` })
		if hrefIndex >= 0 && node.FirstChild != nil {
			var label string
			if node.FirstChild.Type == html.TextNode {
				label = node.FirstChild.Data
			} else if node.FirstChild.DataAtom == atom.Code {
				label = node.FirstChild.FirstChild.Data
			}
			// TODO 需要解 URL 码吗？
			if label != node.Attr[hrefIndex].Val {
				buf.WriteString(label)
				return ast.WalkSkipChildren
			}
		}
		buf.WriteString(`[链接]`)
		return ast.WalkSkipChildren
	},
	`div`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		for _, a := range node.Attr {
			if strings.ToLower(a.Key) == `class` {
				if strings.Contains(a.Val, `footnotes`) {
					return ast.WalkSkipChildren
				} else if strings.Contains(a.Val, `audio-player`) {
					// TODO 用更好的方式，不要特殊处理。
					buf.WriteString(`[音乐]`)
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
	`iframe`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		for _, a := range node.Attr {
			if strings.ToLower(a.Key) == `src` {
				if u, err := url.Parse(a.Val); err == nil {
					switch strings.ToLower(u.Hostname()) {
					case `www.youtube.com`:
						buf.WriteString(`[油管]`)
						return ast.WalkSkipChildren
					}
				}
			}
		}
		buf.WriteString(`[页面]`)
		return ast.WalkSkipChildren
	},
	`sub`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		return ast.WalkSkipChildren
	},
	`sup`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		return ast.WalkSkipChildren
	},
	`span`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		for _, a := range node.Attr {
			if a.Key == `class` && strings.Contains(a.Val, `katex`) {
				buf.WriteString(`[公式]`)
				return ast.WalkSkipChildren
			}
		}
		return ast.WalkContinue
	},
}

type _ContentPrettifier struct{}

func (*_ContentPrettifier) PrettifyHtml(doc *html.Node) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	emojiWithAlt := func(node *html.Node) (alt string) {
		if node.Data != `img` {
			return ""
		}
		img := goquery.NewDocumentFromNode(node)
		if img.HasClass(`emoji`) {
			alt = img.AttrOr(`alt`, ``)
		}
		return
	}

	var walk func(buf *bytes.Buffer, node *html.Node)
	walk = func(buf *bytes.Buffer, node *html.Node) {
		walkStatus := ast.WalkContinue
		switch node.Type {
		case html.TextNode:
			buf.WriteString(node.Data)
		case html.ElementNode:
			walkStatus = ast.WalkContinue
			if f, ok := prettifierStrings[node.Data]; ok {
				// 简短的代码直接显示。
				if node.Data == `code` &&
					node.FirstChild != nil &&
					node.FirstChild.Type == html.TextNode &&
					node.FirstChild.NextSibling == nil &&
					len(node.FirstChild.Data) <= 16 {
					buf.WriteString(node.FirstChild.Data)
					walkStatus = ast.WalkSkipChildren
				} else if alt := emojiWithAlt(node); alt != "" {
					buf.WriteString(alt)
					walkStatus = ast.WalkSkipChildren
				} else {
					buf.WriteString(fmt.Sprintf(`[%s]`, f))
					walkStatus = ast.WalkSkipChildren
				}
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
func New() *_ContentPrettifier {
	return &_ContentPrettifier{}
}
