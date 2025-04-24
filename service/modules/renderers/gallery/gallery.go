package gallery

import (
	"log"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Gallery struct {
}

func New() *Gallery {
	return &Gallery{}
}

func (g *Gallery) TransformHtml(doc *goquery.Document) error {
	doc.Find(`gallery`).Each(func(i int, s *goquery.Selection) {
		replaced, err := g.single(s)
		if err != nil {
			log.Println(err)
			return
		}
		s.ReplaceWithSelection(replaced)
	})
	return nil
}

func (g *Gallery) single(s *goquery.Selection) (*goquery.Selection, error) {
	children := s.Find(`img`)

	div := html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Div,
		Data:     `div`,
		Attr: []html.Attribute{
			{
				Key: `class`,
				Val: `gallery`,
			},
		},
	}

	doc := goquery.NewDocumentFromNode(&div)
	doc.AppendSelection(children)

	return doc.Selection, nil
}
