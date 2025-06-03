package gallery

import (
	"embed"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//go:embed style.css
var _embed embed.FS
var _local = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `gallery`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
	})
}

type Gallery struct{}

func New() *Gallery {
	return &Gallery{}
}

func (g *Gallery) TransformHtml(doc *goquery.Document) error {
	// 旧版本代码。
	doc.Find(`gallery`).Each(func(i int, s *goquery.Selection) {
		replaced, err := g.single(s)
		if err != nil {
			log.Println(err)
			return
		}
		s.ReplaceWithSelection(replaced)
	})

	doc.Find(`p > img:first-child`).Each(func(i int, s *goquery.Selection) {
		elem := s.Nodes[0]
		// 不是第一个元素。
		if elem.PrevSibling != nil {
			return
		}
		// 单张图片，不处理。
		if elem.NextSibling == nil {
			return
		}

		//静态资源。
		if s.HasClass(`static`) {
			return
		}

		// 如果所有兄弟节点都是 <img>
		for ; elem != nil; elem = elem.NextSibling {
			// 换行符不算
			if elem.Type == html.TextNode && strings.TrimSpace(elem.Data) == "" {
				continue
			}
			if elem.DataAtom != atom.Img {
				return
			}
		}

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
		p := s.Parent()
		doc.AppendSelection(p.Children())
		p.ReplaceWithSelection(doc.Selection)
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
