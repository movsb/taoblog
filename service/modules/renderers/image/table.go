package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// TODO 移动到 service/modules/renderers/tables/

// 基于原始JS的Go语言重写版本。
//
// 重写的目的是不想在前端再次引入表格编辑js文件，且直接是后端渲染。
//
// https://github.com/movsb/javascript-table-editor/blob/main/table.js#L512

type TableSchema struct {
	Version int `json:"version"`
	Rows    []struct {
		Cols []struct {
			Data     string `json:"data"`
			Header   bool   `json:"header"`
			Selected bool   `json:"selected"`
			RowSpan  int    `json:"row_span"`
			ColSpan  int    `json:"col_span"`
		} `json:"cols"`
	} `json:"rows"`
}

func renderTableFromJSON(w io.Writer, r io.Reader) {
	decoder := json.NewDecoder(r)
	// 防止不兼容的版本未同步处理。
	decoder.DisallowUnknownFields()

	var schema TableSchema
	if err := decoder.Decode(&schema); err != nil {
		gold_utils.RenderError(w, err)
		return
	}

	if schema.Version > 1 {
		gold_utils.RenderError(w, errors.New("不兼容的表格版本。"))
		return
	}

	table := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Table,
		Data:     `table`,
		Attr: []html.Attribute{
			// 表格是通过img方式插入的，总是会被包裹一层 <p> 标签，导致有语义问题，
			// 这里作一个特殊标记，后面在 TransformHtml 中去除 <p> 标签。
			//
			// 但是！
			//
			// 由于 <p><table> 没有被解析成 <p></p><table>，所以……
			// p > table 其实就已经能找到这个表了。class 不再有意义。
			{
				Key: `class`,
				Val: `editor`,
			},
		},
	}

	body := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Tbody,
		Data:     `tbody`,
	}

	for _, row := range schema.Rows {
		tr := &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Tr,
			Data:     `tr`,
		}

		for _, col := range row.Cols {
			cell := &html.Node{
				Type:     html.ElementNode,
				DataAtom: utils.IIF(col.Header, atom.Th, atom.Td),
				Data:     utils.IIF(col.Header, `th`, `td`),
			}

			cell.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: col.Data,
			})

			if col.Selected {
				cell.Attr = append(cell.Attr, html.Attribute{
					Key: `class`,
					Val: `selected`,
				})
			}

			if col.RowSpan > 1 {
				cell.Attr = append(cell.Attr, html.Attribute{
					Key: `rowspan`,
					Val: fmt.Sprint(col.RowSpan),
				})
			}
			if col.ColSpan > 1 {
				cell.Attr = append(cell.Attr, html.Attribute{
					Key: `colspan`,
					Val: fmt.Sprint(col.ColSpan),
				})
			}

			tr.AppendChild(cell)
		}
		body.AppendChild(tr)
	}

	table.AppendChild(body)

	if err := html.Render(w, table); err != nil {
		gold_utils.RenderError(w, err)
		return
	}
}

// renderImage 在被调用的时候已经处于 <p> 标签内了……
// 对于不能出现在 <p> 标签内的元素，应该想办法去除。
// 比如： ![](1.table) -> <p><table>...</table></p> 是不合法的 HTML。
//
// 奇怪的是，<p><table>...</table></p> 对于 html.Parse 居然是合法的，
// 我还以为 table 不能在 p 中，然后被 parse 成 <p></p><table>...</table><p></p>……
func fixImageTable(doc *goquery.Document) {
	doc.Find(`p > table.editor`).Each(func(i int, s *goquery.Selection) {
		s.Parent().ReplaceWithSelection(s)
	})
}
