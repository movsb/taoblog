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
	Text    string   `yaml:"text"`
	ColSpan int      `yaml:"colspan"`
	RowSpan int      `yaml:"rowspan"`
	Formats []string `yaml:"formats"`
	Styles  Styles   `yaml:"styles"`
}

type _PlainCol Col

func (c *Col) UnmarshalYAML(unmarshal func(any) error) error {
	var err error
	if err = unmarshal(&c.Text); err == nil {
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
	buf.WriteString(html.EscapeString(c.Text))
	buf.WriteString(`</td>`)
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
