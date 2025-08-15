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
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
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

	// 不要求每行字段数一致，后面会自动补齐。
	csvReader.FieldsPerRecord = -1

	table, err := csvReader.ReadAll()
	if err != nil {
		gold_utils.RenderError(w, fmt.Errorf(`failed to parse csv as table: %w`, err))
		return nil
	}

	if len(table) <= 0 {
		return nil
	}

	// 正在编辑时可能出现列数不一致的情况，没必要报错。
	// 采用自动填充空数据的方式手动对齐。
	m := 0
	for _, cols := range table {
		m = max(m, len(cols))
	}
	for i, cols := range table {
		d := m - len(cols)
		for range d {
			table[i] = append(table[i], ``)
		}
	}

	return ttt().Execute(w, table)
}
