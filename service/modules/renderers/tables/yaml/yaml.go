package yaml

import (
	"fmt"
	"html"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
)

type Table struct {
	// 存放用于被引用的 Anchor 数据。
	Cells any `yaml:"cells"`

	// 包含两个元素的数组，分别表示哪些行、哪些列作为表头渲染，从 1 开始。
	Headers [2]Headers `yaml:"headers"`

	// 行数据。每行包含所有的列数据。
	Rows []*Row `yaml:"rows"`

	// 临时：列格式控制。
	Cols map[int]struct {
		Formats []string `yaml:"formats"`
	} `yaml:"cols"`

	// <table>/<th>/<td> 的边框设置。
	Border Border `yaml:"border"`

	headerRows map[int]struct{}
	headerCols map[int]struct{}

	markdownRenderer MarkdownRenderer
	outputCoords     bool
}

type Border struct {
	b *bool
}

func (b *Border) UnmarshalYAML(unmarshal func(any) error) error {
	var bb bool
	if err := unmarshal(&bb); err == nil {
		b.b = &bb
		return nil
	}
	return fmt.Errorf(`bad border`)
}

type Headers []int

func (h *Headers) UnmarshalYAML(unmarshal func(any) error) error {
	if err := unmarshal((*[]int)(h)); err == nil {
		return nil
	}
	var n int
	if err := unmarshal(&n); err == nil {
		*h = []int{n}
		return nil
	}
	return fmt.Errorf(`bad headers`)
}

type MarkdownRenderer func(text string) (string, error)

func (t *Table) SetTextRenderer(tr MarkdownRenderer) {
	t.markdownRenderer = tr
}

func (t *Table) SetOutputCoords(coords bool) {
	t.outputCoords = coords
}

func (t *Table) isTH(r, c int) bool {
	_, ok1 := t.headerRows[r]
	_, ok2 := t.headerCols[c]
	return ok1 || ok2
}

func (t *Table) setRoot() {
	for i, row := range t.Rows {
		if row == nil {
			r := &Row{}
			t.Rows[i] = r
			row = r
		}
		row.root = t
		for j, col := range row.Cols {
			if col == nil {
				c := &Col{}
				row.Cols[j] = c
				col = c
			}
			col.setRoot(t)
		}
	}
}

func (t *Table) calculateCoords() {
	if len(t.Headers) >= 1 {
		t.headerRows = make(map[int]struct{})
		for _, n := range t.Headers[0] {
			t.headerRows[n] = struct{}{}
		}
	}
	if len(t.Headers) >= 2 {
		t.headerCols = make(map[int]struct{})
		for _, n := range t.Headers[1] {
			t.headerCols[n] = struct{}{}
		}
	}

	for r, row := range t.Rows {
		for c, col := range row.Cols {
			p := &col.coords

			// (r,c) 是物理位置。

			// 行始终是不变的。
			p.r1 = r + 1

			// for debug
			// if r == 2 && c == 0 {
			// 	r += 0
			// }

			if c > 0 {
				// 列会被左边的往右边挤。
				// 所以计算的时候要参考左边是多少。
				right := row.Cols[c-1].coords.c2
				p.c1 = right + 1
			} else {
				// 但是左边可能会是上边的 rowspan 挤下来的，物理上不一定相邻。
				if r == 0 {
					p.c1 = 1
				} else {
					// 根据物理位置依次判断物理坐标是否被前面的元素占领。
					tr, tc := r+1, c+1
				retry:
					for x := 0; x <= r; x++ {
						for y := 0; y <= c; y++ {
							if x == r && y == c {
								goto out
							}
							cc := t.Rows[x].Cols[y].coords
							if cc.includes(tr, tc) {
								tc++
								goto retry
							}
						}
					}
				out:
					p.c1 = tc
				}
			}

			if col.RowSpan == 0 {
				p.r2 = p.r1
			} else {
				p.r2 = p.r1 + col.RowSpan - 1
			}

			if col.ColSpan == 0 {
				p.c2 = p.c1
			} else {
				p.c2 = p.c1 + col.ColSpan - 1
			}
		}
	}
}

type _PlainTable Table

func (t *Table) UnmarshalYAML(unmarshal func(any) error) error {
	if err := unmarshal(&t.Rows); err == nil {
		return nil
	}
	return unmarshal((*_PlainTable)(t))
}

func (t *Table) Render(buf *strings.Builder) error {
	t.setRoot()
	t.calculateCoords()

	buf.WriteString(`<table class="yaml`)
	if b := t.Border.b; b != nil {
		buf.WriteString(utils.IIF(*b, ` border`, ` no-border`))
	}
	buf.WriteString("\">\n")

	// 暂时没使用。
	// 表头更应该放在 thead 中，方便打印时自动分页的时候
	// 每个续表都自动添加表头。
	buf.WriteString("<thead>\n")
	buf.WriteString("</thead>\n")

	buf.WriteString("<tbody>\n")

	for _, row := range t.Rows {
		if err := row.Render(buf); err != nil {
			return err
		}
	}

	buf.WriteString("</tbody>\n")
	buf.WriteString("</table>\n")
	return nil
}

type Row struct {
	root *Table

	Cols    []*Col   `yaml:"cols"`
	Formats []string `yaml:"formats"`
}

func (r *Row) Render(buf *strings.Builder) error {
	buf.WriteString(`<tr`)
	if len(r.Formats) > 0 {
		buf.WriteString(` class="`)
		renderFormats(buf, r.Formats)
		buf.WriteString(`"`)
	}
	buf.WriteString(`>`)
	for _, col := range r.Cols {
		if err := col.Render(buf); err != nil {
			return err
		}
	}
	buf.WriteString("</tr>\n")
	return nil
}

type _PlainRow Row

func (r *Row) UnmarshalYAML(unmarshal func(any) error) error {
	if err := unmarshal(&r.Cols); err == nil {
		return nil
	}
	return unmarshal((*_PlainRow)(r))
}

type Col struct {
	root *Table

	// 文本、Markdown、表格或是数组构成的。
	// 四选一。
	Text     string `yaml:"text"`
	Markdown string `yaml:"markdown"`
	Table    *Table `yaml:"table"`
	Cols     []*Col

	// 值为 0 时表示没设置，处理为 1.
	ColSpan int `yaml:"colspan"`
	RowSpan int `yaml:"rowspan"`

	Formats []string `yaml:"formats"`
	// Styles  Styles   `yaml:"styles"`

	coords Coords
}

// 逻辑坐标。
// 左上角为(1,1)，包含右下角。
// 主要用于确定一个单元格在显示上属于第几行、第几列。
type Coords struct {
	r1, c1 int
	r2, c2 int
}

func (cc Coords) includes(r, c int) bool {
	return cc.r1 <= r && r <= cc.r2 && cc.c1 <= c && c <= cc.c2
}

type _PlainCol Col

func (c *Col) UnmarshalYAML(unmarshal func(any) error) error {
	var err error
	if err = unmarshal(&c.Text); err == nil {
		return nil
	}
	if err := unmarshal(&c.Cols); err == nil {
		return nil
	}
	if err = unmarshal((*_PlainCol)(c)); err == nil {
		return nil
	}
	return fmt.Errorf(`do not know how to unmarshal to Col: %w`, err)
}

func (c *Col) Render(buf *strings.Builder) error {
	isTH := c.root.isTH(c.coords.r1, c.coords.c1)
	buf.WriteString(utils.IIF(isTH, `<th`, `<td`))

	if c.ColSpan > 1 {
		fmt.Fprintf(buf, ` colspan="%d"`, c.ColSpan)
	}
	if c.RowSpan > 1 {
		fmt.Fprintf(buf, ` rowspan="%d"`, c.RowSpan)
	}

	if c.root.outputCoords {
		buf.WriteString(` data-coords="`)
		fmt.Fprintf(buf, `[%d,%d,%d,%d]`, c.coords.r1, c.coords.c1, c.coords.r2, c.coords.c2)
		buf.WriteString(`"`)
	}

	colAttr, _ := c.root.Cols[c.coords.c1]

	if len(c.Formats) > 0 || len(colAttr.Formats) > 0 {
		buf.WriteString(` class="`)
		all := append([]string{}, c.Formats...)
		all = append(all, colAttr.Formats...)
		if err := renderFormats(buf, all); err != nil {
			return err
		}
		buf.WriteString(`"`)
	}

	buf.WriteByte('>')

	if err := c.renderContent(buf); err != nil {
		return err
	}

	buf.WriteString(utils.IIF(isTH, `</th>`, `</td>`))
	return nil
}

// t := fmt.Sprintf(`(%d,%d)(%d,%d)`, c.coords.r1, c.coords.c1, c.coords.r2, c.coords.c2)
// writeText(buf, t)
func writeText(buf *strings.Builder, t string) {
	if strings.ContainsAny(t, " \n\t") {
		buf.WriteString(`<span class="pre">`)
		defer buf.WriteString(`</span>`)

	}
	buf.WriteString(html.EscapeString(t))
}

func (c *Col) renderContent(buf *strings.Builder) error {
	if c.Text != `` {
		writeText(buf, c.Text)
	} else if c.Markdown != `` {
		if tr := c.root.markdownRenderer; tr != nil {
			text, err := tr(c.Markdown)
			if err != nil {
				return err
			}
			buf.WriteString(text)
		} else {
			writeText(buf, c.Markdown)
		}
	} else if c.Table != nil {
		if err := c.Table.Render(buf); err != nil {
			return err
		}
	} else if len(c.Cols) > 0 {
		for i := 0; i < len(c.Cols); {
			// 如果有其它块级元素，则文本元素应该放在独立的块中。
			// 这里是为了合并多个文本。
			j := i
			for j < len(c.Cols) && c.Cols[j].Table == nil && len(c.Cols[j].Cols) == 0 {
				j++
			}
			if j > i {
				buf.WriteString(`<p>`)
				for k := i; k < j; k++ {
					kc := c.Cols[k]
					if len(kc.Formats) > 0 {
						buf.WriteString(`<span class="`)
						if err := renderFormats(buf, kc.Formats); err != nil {
							return err
						}
						buf.WriteString(`">`)
					}
					if err := kc.renderContent(buf); err != nil {
						return err
					}
					if len(kc.Formats) > 0 {
						buf.WriteString(`</span>`)
					}
				}
				buf.WriteString(`</p>`)
				i = j
			} else {
				if err := c.Cols[i].renderContent(buf); err != nil {
					return err
				}
				i++
			}
		}
	}
	return nil
}

func (c *Col) setRoot(t *Table) {
	c.root = t
	for _, col := range c.Cols {
		col.setRoot(t)
	}
	if c.Table != nil {
		c.Table.setRoot()
	}
}

func renderFormats(buf *strings.Builder, formats []string) error {
	for i, format := range formats {
		if i > 0 {
			buf.WriteByte(' ')
		}
		switch format {
		case `bold`, `italic`, `strike`, `underline`, `code`, `kbd`, `ins`, `del`, `left`, `center`, `right`:
			buf.WriteString(format)
		default:
			return fmt.Errorf(`unknown format: %s`, format)
		}
	}
	return nil
}
