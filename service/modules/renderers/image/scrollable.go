package image

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (e *Image) TransformHtml(doc *goquery.Document) error {
	doc.Find(`p > img:only-child, p > picture:only-child`).Each(func(i int, s *goquery.Selection) {
		// 可能有文本节点，排除。
		elem := s.Nodes[0]
		if elem.PrevSibling != nil || elem.NextSibling != nil {
			return
		}

		div := &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Div,
			Data:     `div`,
			Attr: []html.Attribute{
				{
					Key: `class`,
					Val: `image-scroll-outer`,
				},
			},
		}

		s.Parent().WrapNode(div)
	})
	return nil
}
