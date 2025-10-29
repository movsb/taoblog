package blur_image

import (
	"embed"
	"log"
	urlpkg "net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
)

//go:embed thumbhash/hash.js script.js
var _embed embed.FS
var _root = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `blur_image`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithScripts(module, `thumbhash/hash.js`, `script.js`)
	})
}

func New(web gold_utils.WebFileSystem) any {
	return &BlurImage{
		web: web,
	}
}

type BlurImage struct {
	web gold_utils.WebFileSystem
}

func (m *BlurImage) TransformHtml(doc *goquery.Document) error {
	images := doc.Find(`img:not(.static)`)
	images.Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr(`src`, ``)
		if url == `` {
			return
		}

		parsedURL, err := urlpkg.Parse(url)
		if err != nil {
			log.Println(err)
			return
		}
		if parsedURL.Scheme != `` || parsedURL.Host != `` {
			return
		}
		f, err := m.web.OpenURL(url)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			return
		}
		info, ok := stat.Sys().(*models.File)
		if !ok {
			return
		}
		if hash := info.Meta.ThumbHash; hash != `` {
			s.SetAttr(`data-thumb-hash`, hash)
		}
	})
	return nil
}
