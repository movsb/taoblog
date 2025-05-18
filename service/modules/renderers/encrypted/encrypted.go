package encrypted

import (
	"embed"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
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

type _Encrypted struct{}

func New() *_Encrypted {
	return &_Encrypted{}
}

func (m *_Encrypted) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img,video`).Each(func(i int, s *goquery.Selection) {
		s.SetAttr(`onerror`, `decryptFile(this)`)
	})
	return nil
}
