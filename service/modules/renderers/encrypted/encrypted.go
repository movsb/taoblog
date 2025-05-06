package encrypted

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `encrypted`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
		dynamic.WithScripts(module, `script.js`)
	})
}

//go:embed style.css script.js
var _embed embed.FS
var _local = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

type _Encrypted struct {
	web gold_utils.WebFileSystem
}

func New(web gold_utils.WebFileSystem) *_Encrypted {
	return &_Encrypted{
		web: web,
	}
}

func (m *_Encrypted) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img`).Each(func(i int, s *goquery.Selection) {
		src := s.AttrOr(`src`, ``)
		if src == `` {
			return
		}

		fp, err := m.web.OpenURL(src)
		if err != nil {
			return
		}
		defer fp.Close()
		sys, ok := utils.Must1(fp.Stat()).Sys().(*models.File)
		if !ok {
			return
		}

		random := utils.RandomString()
		s.SetAttr(`data-id`, random)
		s.SetAttr(`data-encryption`, string(utils.Must1(json.Marshal(sys.Meta.Encryption))))
		s.SetAttr(`onerror`, fmt.Sprintf(`javascript:decodeFile("%s")`, random))
	})
	return nil
}
