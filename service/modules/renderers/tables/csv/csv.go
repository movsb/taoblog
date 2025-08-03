package csv

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"sync"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/yuin/goldmark/parser"
)

// 只能用逗号作为分隔符。
//
// 第一行总是作为标题。
//
// CSV 中的所有的空白会被保留。
// 因为可以用 " " 来输入空格、换行符等。
//
// 没有任何格式控制支持。如加粗、字体颜色。
//
//   - [Content adapters](https://gohugo.io/content-management/content-adapters/)
//   - [Comma-separated values - Wikipedia](https://en.wikipedia.org/wiki/Comma-separated_values)
//   - [template package - text/template - Go Packages](https://pkg.go.dev/text/template)
type CSV struct{}

func New() *CSV {
	p := &CSV{}

	return p
}

var (
	//go:embed tmpl.html
	tt  []byte
	ttt = sync.OnceValue(func() *template.Template {
		return template.Must(template.New(`csv`).Parse(string(tt)))
	})
)

func (t *CSV) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) (outErr error) {
	defer utils.CatchAsError(&outErr)

	csvReader := csv.NewReader(bytes.NewReader(source))
	table, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf(`failed to parse csv as table: %w`, err)
	}

	if len(table) <= 0 {
		return nil
	}

	return ttt().Execute(w, table)
}
