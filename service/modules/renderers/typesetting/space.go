package typesetting

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Space struct {
}

func NewSpace() *Space {
	return &Space{}
}

func (s *Space) TransformHtml(doc *goquery.Document) error {
	s.replace(doc.Nodes[0])
	return nil
}

func (s *Space) replace(root *html.Node) {

	textNodes := make([]*html.Node, 0, 128)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// https://html.spec.whatwg.org/multipage/syntax.html#elements-2
			switch n.DataAtom {
			case atom.Template:
				// The template element
				return
			case atom.Script, atom.Style:
				// Raw text elements
				return
			case atom.Textarea, atom.Title:
				// Escapable raw text elements
				return
			case atom.Svg, atom.Math:
				// Foreign elements
				return
			case atom.Pre, atom.Code:
				// 额外加的。其实可以不处理。
				return
			}
		}
		if n.Type == html.TextNode {
			// 简单只处理空格(0x20)和换行。
			allSpaces := func(s string) bool {
				for i := 0; i < len(s); i++ {
					if s[i] != 0x20 && s[i] != 0x10 {
						return false
					}
				}
				return true
			}
			if !allSpaces(n.Data) && strings.IndexByte(n.Data, ' ') != -1 {
				textNodes = append(textNodes, n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)

	for _, text := range textNodes {
		s := text.Data
		var parts []string
		for i := 0; i < len(s); {
			j := i
			if s[i] == ' ' {
				for j < len(s) && s[j] == ' ' {
					j++
				}
			} else {
				for j < len(s) && s[j] != ' ' {
					j++
				}
			}
			parts = append(parts, s[i:j])
			i = j
		}
		parent := text.Parent
		for _, part := range parts {
			n := &html.Node{}
			if part[0] == ' ' {
				n.Type = html.ElementNode
				n.DataAtom = atom.Span
				n.Data = `span`
				n.Attr = append(n.Attr, html.Attribute{
					Key: `class`,
					Val: `space`,
				})
				n.AppendChild(&html.Node{
					Type: html.TextNode,
					Data: part,
				})
			} else {
				n.Type = html.TextNode
				n.Data = part
			}
			parent.InsertBefore(n, text)
		}
		parent.RemoveChild(text)
	}
}
