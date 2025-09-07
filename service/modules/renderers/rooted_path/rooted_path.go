package rooted_path

import (
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"golang.org/x/net/html/atom"
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
		case `img`, `iframe`, `audio`, `video`:
			attrName = `src`
		case `object`:
			attrName = `data`
		case `source`:
			// [<source>: The Media or Image Source element - HTML | MDN](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Elements/source#src)
			// <audio> 和 <video> 用 src
			// <picture> 用 srcset
			switch s.Parent().Nodes[0].DataAtom {
			case atom.Audio, atom.Video:
				attrName = `src`
			case atom.Picture:
				attrName = `srcset`
			default:
				// 其实不存在了，测试里面有个没有 parent
				attrName = `src`
			}
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
