package tables

import (
	"embed"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `tables`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

type _TableWrapper struct{}

func (_TableWrapper) TransformHtml(doc *goquery.Document) error {
	doc.Find(`table`).Each(func(i int, s *goquery.Selection) {
		s.WrapHtml(`<div class="table-wrapper"></div>`)
	})
	return nil
}

func NewTableWrapper() gold_utils.HtmlTransformer {
	return &_TableWrapper{}
}
