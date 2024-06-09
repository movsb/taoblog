package rooted_path

import (
	"log"

	"github.com/PuerkitoBio/goquery"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
)

type _RootedPaths struct {
	web gold_utils.WebFileSystem
}

func New(web gold_utils.WebFileSystem) *_RootedPaths {
	return &_RootedPaths{
		web: web,
	}
}

func (m *_RootedPaths) TransformHtml(doc *goquery.Document) error {
	doc.Find(`a,img,iframe,source,audio,video,object`).Each(func(_ int, s *goquery.Selection) {
		var attrName string
		switch s.Nodes[0].Data {
		case `a`:
			attrName = `href`
		case `img`, `iframe`, `source`, `audio`, `video`:
			attrName = `src`
		case `object`:
			attrName = `data`
		}
		attrValue := s.AttrOr(attrName, ``)
		if attrValue == `` {
			return
		}
		ru, err := m.web.Resolve(attrValue, true)
		if err != nil {
			log.Println(err)
			return
		}
		s.SetAttr(attrName, ru.String())
	})
	return nil
}
