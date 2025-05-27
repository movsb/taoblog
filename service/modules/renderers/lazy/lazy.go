package lazy

import "github.com/PuerkitoBio/goquery"

type Lazy struct{}

var lazy = &Lazy{}

func New() *Lazy {
	return lazy
}

func (m *Lazy) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img,iframe`).Each(func(i int, s *goquery.Selection) {
		// 仅针对未设置的元素进行设置。
		_, ok := s.Attr(`loading`)
		if ok {
			return
		}

		s.SetAttr(`loading`, `lazy`)
	})

	doc.Find(`audio,video`).Each(func(i int, s *goquery.Selection) {
		_, ok := s.Attr(`preload`)
		if ok {
			return
		}
		s.SetAttr(`preload`, `none`)
	})

	return nil
}
