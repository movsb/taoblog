package tables

import (
	"bytes"
	"embed"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/tables/csv"
	yt "github.com/movsb/taoblog/service/modules/renderers/tables/yaml"
	"github.com/yuin/goldmark/parser"
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
		// TODO 临时过滤。
		if s.Parent().HasClass(`chroma`) {
			return
		}
		s.WrapHtml(`<div class="table-wrapper"></div>`)
	})
	return nil
}

func NewTableWrapper() gold_utils.HtmlTransformer {
	return &_TableWrapper{}
}

func NewCSV() *csv.CSV {
	return csv.New()
}

func NewYAML(tr yt.MarkdownRenderer) *_YAML {
	return &_YAML{tr: tr}
}

type _YAML struct {
	tr yt.MarkdownRenderer
}

func (t *_YAML) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) (outErr error) {
	defer utils.CatchAsError(&outErr)

	var table yt.Table
	utils.Must(yaml.NewDecoder(bytes.NewReader(source), yaml.DisallowUnknownField()).Decode(&table))

	table.SetTextRenderer(t.tr)

	buf := strings.Builder{}
	utils.Must(table.Render(&buf))
	_, err := w.Write([]byte(buf.String()))
	// log.Println(buf.String())
	return err
}
