package tables

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
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

func (t *_YAML) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) error {
	var table yt.Table
	if err := yaml.NewDecoder(bytes.NewReader(source), yaml.DisallowUnknownField()).Decode(&table); err != nil {
		if errors.Is(err, io.EOF) {
			err = fmt.Errorf(`Error: empty YAML table.`)
		}
		gold_utils.RenderError(w, err)
		return nil
	}

	table.SetTextRenderer(t.tr)

	buf := strings.Builder{}
	if err := table.Render(&buf); err != nil {
		gold_utils.RenderError(w, err)
		return nil
	}

	_, err := w.Write([]byte(buf.String()))
	return err
}
