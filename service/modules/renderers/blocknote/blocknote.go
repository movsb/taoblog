package blocknote

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
)

type Blocknote struct {
	title            *string
	doNotRenderTitle bool
}

func New(options ...Option) *Blocknote {
	b := &Blocknote{}
	for _, opt := range options {
		opt(b)
	}
	return b
}

func (b *Blocknote) Render(source string) (string, error) {
	blocks := []*Block{}

	dec := json.NewDecoder(strings.NewReader(source))
	dec.UseNumber()
	dec.DisallowUnknownFields()
	if err := dec.Decode(&blocks); err != nil {
		return ``, fmt.Errorf(`blocknote: error rendering: %w`, err)
	}

	buf := strings.Builder{}
	b.renderBlocks(&buf, blocks)

	return buf.String(), nil
}

func (b *Blocknote) renderBlocks(buf *strings.Builder, blocks []*Block) {
	for _, block := range blocks {
		b.renderBlock(buf, block, true)
		b.renderBlock(buf, block, false)
	}
}

func (b *Blocknote) renderBlock(buf *strings.Builder, block *Block, entering bool) {
	switch block.Type {
	default:
		panic(`未知标签类型：` + block.Type)
	case `heading`:
		levelString, _ := block.Props[`level`].(json.Number)
		level := utils.Must1(levelString.Int64())
		if entering {
			fmt.Fprintf(buf, `<h%d>`, level)

			if level == 1 && !b.doNotRenderTitle {
				b.renderContent(buf, block.Content)
			}

			if level == 1 && b.title != nil {
				if *b.title != `` {
					panic(`重复标题`)
				}
				// TODO: 重复 render 了，确保 b 没有状态。
				t := strings.Builder{}
				b.renderContent(&t, block.Content)
				*b.title = t.String()
			}
		} else {
			fmt.Fprintf(buf, `</h%d>`, level)
		}
	case `quote`:
		if entering {
			buf.WriteString(`<blockquote`)
			writeProps(buf, block.Props)
			buf.WriteString(`>`)
			b.renderContent(buf, block.Content)
		} else {
			buf.WriteString(`</blockquote>`)
		}
	case `paragraph`:
		if entering {
			buf.WriteString(`<p`)
			writeProps(buf, block.Props)
			buf.WriteString(`>`)
			b.renderContent(buf, block.Content)
		} else {
			buf.WriteString(`</p>`)
		}
	case `toggleListItem`:
		if entering {
			buf.WriteString(`<details>`)
			buf.WriteString(`<summary`)
			writeProps(buf, block.Props)
			buf.WriteByte('>')
			b.renderContent(buf, block.Content)
			buf.WriteString(`</summary>`)
			b.renderBlocks(buf, block.Children)
		} else {
			buf.WriteString(`</details>`)
		}
	case `table`:
		if entering {
			buf.WriteString(`<table>`)
			b.renderContent(buf, block.Content)
		} else {
			buf.WriteString(`</table>`)
		}
	}
}

func writeProps(buf *strings.Builder, props map[string]any) {
	styles := strings.Builder{}
	attrs := strings.Builder{}
	for key, val := range props {
		switch key {
		case `textColor`:
			s, _ := val.(string)
			if s != `default` {
				fmt.Fprintf(&styles, `color:%s;`, s)
			}
		case `backgroundColor`:
			s, _ := val.(string)
			if s != `default` {
				fmt.Fprintf(&styles, `background-color:%s;`, s)
			}
		case `textAlignment`:
			if v := val.(string); v != `left` {
				fmt.Fprintf(&styles, `text-align:%s;`, v)
			}
		case `colspan`, `rowspan`:
			v, _ := val.(json.Number).Int64()
			if v > 1 {
				fmt.Fprintf(&attrs, ` %s=%d`, key, v)
			}
		}
	}
	if styles.Len() > 0 {
		buf.WriteString(` style="`)
		buf.WriteString(styles.String())
		buf.WriteString(`"`)
	}
	if attrs.Len() > 0 {
		buf.WriteString(attrs.String())
	}
}

func (b *Blocknote) renderContent(buf *strings.Builder, content *Content) {
	if len(content.Inlines) > 0 {
		b.renderInlines(buf, content.Inlines)
		return
	}
	if content.Object != nil {
		switch typed := content.Object.(type) {
		default:
			panic(`unknown content object type`)
		case *TableContent:
			b.renderTableContent(buf, typed)
			return
		}
	}
}

func (b *Blocknote) renderTableContent(buf *strings.Builder, t *TableContent) {
	buf.WriteString(`<tbody>`)

	for i, row := range t.Rows {
		headerRow := t.HeaderRows > 0 && i+1 == t.HeaderRows
		buf.WriteString(`<tr>`)
		for j, col := range row.Cells {
			headerCol := t.HeaderCols > 0 && j+1 == t.HeaderCols
			b.renderTableCell(buf, col, headerRow || headerCol)
		}
		buf.WriteString(`</tr>`)
	}

	buf.WriteString(`</tbody>`)
}

func (b *Blocknote) renderTableCell(buf *strings.Builder, block *Block, header bool) {
	buf.WriteString(utils.IIF(header, `<th`, `<td`))
	writeProps(buf, block.Props)
	buf.WriteString(`>`)
	b.renderContent(buf, block.Content)
	buf.WriteString(utils.IIF(header, `</th>`, `</td>`))
}

func (b *Blocknote) renderInlines(buf *strings.Builder, inlines []*Inline) {
	for _, inline := range inlines {
		b.renderInline(buf, inline, true)
		b.renderInline(buf, inline, false)
	}
}

func (b *Blocknote) renderInline(buf *strings.Builder, inline *Inline, entering bool) {
	switch inline.Type {
	default:
		panic(`未知标签类型：` + inline.Type)
	case `text`:
		if entering {
			var (
				hasBold          bool
				hasItalic        bool
				hasUnderline     bool
				hasStrike        bool
				fgColor, bgColor string
			)

			for style, value := range inline.Styles {
				switch style {
				case `bold`:
					hasBold, _ = value.(bool)
				case `italic`:
					hasItalic, _ = value.(bool)
				case `underline`:
					hasUnderline, _ = value.(bool)
				case `strike`:
					hasStrike, _ = value.(bool)
				case `textColor`:
					fgColor = value.(string)
				case `backgroundColor`:
					bgColor = value.(string)
				}
			}

			hasColor := fgColor != `` || bgColor != ``

			withCond(hasBold,
				func() { buf.WriteString(`<strong>`) },
				func() { buf.WriteString(`</strong>`) },
				withCond(hasItalic,
					func() { buf.WriteString(`<em>`) },
					func() { buf.WriteString(`</em>`) },
					withCond(hasUnderline,
						func() { buf.WriteString(`<u>`) },
						func() { buf.WriteString(`</u>`) },
						withCond(hasStrike,
							func() { buf.WriteString(`<s>`) },
							func() { buf.WriteString(`</s>`) },
							withCond(hasColor,
								func() {
									buf.WriteString(`<span style="`)
									if fgColor != `` {
										fmt.Fprintf(buf, `color: %s;`, fgColor)
									}
									if bgColor != `` {
										fmt.Fprintf(buf, `background-color: %s;`, bgColor)
									}
									buf.WriteString(`">`)
								},
								func() {
									buf.WriteString(`</span>`)
								},
								func() {

									buf.WriteString(html.EscapeString(inline.Text))
								},
							),
						),
					),
				),
			)()
		} else {
		}
	}
}

func withCond(c bool, pre func(), post func(), real func()) func() {
	return func() {
		if c {
			pre()
		}
		real()
		if c {
			post()
		}
	}
}
