package renderers

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// 简化并美化 Markdown 的展示。
// 比如：
// - 不显示复杂的链接、图片、表格、代码块等元素。
// - 不显示脚注。
type Prettifier struct {
}

func (p *Prettifier) Prettify(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)

	var walk func(buf *bytes.Buffer, node *html.Node)
	walk = func(buf *bytes.Buffer, node *html.Node) {
		switch node.Type {
		case html.DocumentNode:
		case html.TextNode:
			buf.WriteString(node.Data)
		case html.ElementNode:
			switch node.Data {
			case `a`:
				sub := bytes.NewBuffer(nil)
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					walk(sub, c)
				}
				buf.WriteString(fmt.Sprintf(`[链接:%s]`, sub.String()))
				return
			case `img`:
				var alt string
				for _, a := range node.Attr {
					if strings.ToLower(a.Key) == `alt` {
						alt = a.Val
					}
				}
				if alt == "" {
					alt = "图片"
				}
				buf.WriteString(fmt.Sprintf(`[图片:%s]`, alt))
			case `div`:
				for _, a := range node.Attr {
					if strings.ToLower(a.Key) == `class` {
						if strings.Contains(a.Val, `footnotes`) {
							return
						}
					}
				}
			case `pre`:
				if node.FirstChild != nil && node.FirstChild.NextSibling == nil && node.FirstChild.Data == `code` {
					buf.WriteString(`[代码]`)
					return
				}
			case `table`:
				buf.WriteString(`[表格]`)
				return
			}
		default:
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(buf, c)
		}
	}

	walk(buf, doc)
	return buf.String(), nil
}
