package encrypted

import (
	"embed"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
)

func init() {
	dynamic.RegisterInit(func() {
		const module = `encrypted`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithScripts(module, `script.js`)
	})
}

//go:embed script.js
var _embed embed.FS
var _local = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

type _Encrypted struct{}

func New() *_Encrypted {
	return &_Encrypted{}
}

func (m *_Encrypted) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img,video`).Each(func(i int, s *goquery.Selection) {
		if s.HasClass(`static`) {
			return
		}
		s.SetAttr(`onerror`, `decryptFile(this)`)
	})
	return nil
}
