package auto_image_border

import (
	"embed"
	"fmt"
	"log"
	urlpkg "net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `auto_image_border`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

func New(web gold_utils.WebFileSystem) any {
	return &AutoImageBorder{
		web: web,
	}
}

type AutoImageBorder struct {
	web gold_utils.WebFileSystem
}

func (m *AutoImageBorder) TransformHtml(doc *goquery.Document) error {
	images := doc.Find(`p>img:not(.static):only-child`)
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
		if r := info.Meta.BorderContrastRatio; r > 0 {
			s.SetAttr(`data-contrast-ratio`, fmt.Sprint(info.Meta.BorderContrastRatio))
			if r < 0.4 {
				s.AddClass(`border`)
			}
		}
	})
	return nil
}
