package yaml

import (
	"fmt"
	"html"
	"strings"
)

type Table struct {
	Rows []Row `yaml:"rows"`
}

func (t *Table) Render(buf *strings.Builder) error {
	buf.WriteString("<table>\n")
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
	Cols []Col `yaml:"cols"`
}

func (r *Row) Render(buf *strings.Builder) error {
	buf.WriteString(`<tr>`)
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
	// 文本、表格或是数组构成的。
	// 三选一。
	Text  string `yaml:"text"`
	Table *Table `yaml:"table"`
	Cols  []*Col

	ColSpan int `yaml:"colspan"`
	RowSpan int `yaml:"rowspan"`

	Formats []string `yaml:"formats"`
	// Styles  Styles   `yaml:"styles"`
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
	buf.WriteString(`<td`)
	if c.ColSpan > 1 {
		fmt.Fprintf(buf, ` colspan="%d"`, c.ColSpan)
	}
	if c.RowSpan > 1 {
		fmt.Fprintf(buf, ` rowspan="%d"`, c.RowSpan)
	}
	if len(c.Formats) > 0 {
		buf.WriteString(` class="`)
		if err := renderFormats(buf, c.Formats); err != nil {
			return err
		}
		buf.WriteString(`"`)
	}
	buf.WriteByte('>')

	if err := c.renderContent(buf); err != nil {
		return err
	}

	buf.WriteString(`</td>`)
	return nil
}

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

type Styles struct {
	Color      string `yaml:"color"`
	Background string `yaml:"background"`
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

// func renderStyles(buf *strings.Builder, styles Styles) error {
// 	sb := strings.Builder{}
// 	if styles.Color != `` {

// 	}
// }
