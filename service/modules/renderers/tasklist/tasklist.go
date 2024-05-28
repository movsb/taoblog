package task_list

import (
	"fmt"
	"regexp"
	"strings"

	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// 任务/待办列表（TODO list）。
//
// 支持标记任务在原文中的位置，以方便在页面上点击后直接修改任务状态。
//
// 格式：
//
//	<!-- TaskList -->
//
// - [ ] Task1
// - [x] Task2
func New() *TaskList {
	return &TaskList{}
}

type TaskList struct{}

var _ interface {
	goldmark.Extender
	parser.ASTTransformer
	renderer.NodeRenderer
} = (*TaskList)(nil)

func (e *TaskList) Extend(m goldmark.Markdown) {
	extension.TaskList.Extend(m)
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(e, 100),
	))
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(e, 100),
	))
}

var reIsTaskListComment = regexp.MustCompile(`(?i:^\s{0,3}<!--\s*TaskList\s*-->\s*$)`)

// ... implements parser.ASTTransformer.
// TODO 只处理第一层任务。
func (e *TaskList) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	recurse := func(list *ast.List) {
		gold_utils.AddClass(list, `task-list`)
		for c := list.FirstChild(); c != nil; c = c.NextSibling() {
			item, ok := c.(*ast.ListItem)
			if !ok {
				continue
			}
			if item.FirstChild().Kind() == ast.KindTextBlock {
				tb := item.FirstChild().(*ast.TextBlock)
				if tb.FirstChild() == nil || tb.FirstChild().Kind() != extast.KindTaskCheckBox {
					continue
				}
				checkBox := tb.FirstChild().(*extast.TaskCheckBox)
				classes := []string{`task-list-item`}
				if checkBox.IsChecked {
					classes = append(classes, `checked`)
				}
				gold_utils.AddClass(item, classes...)
				item.SetAttributeString(`data-source-position`, fmt.Sprint(checkBox.Segment.Start))
			}
		}
	}

	ast.Walk(node, func(n ast.Node, entering bool) (status ast.WalkStatus, err error) {
		status = ast.WalkContinue
		if !entering {
			return
		}
		if n.Kind() != ast.KindHTMLBlock {
			return
		}
		hb := n.(*ast.HTMLBlock)
		// https://spec.commonmark.org/0.31.2/#html-block
		if hb.HTMLBlockType != ast.HTMLBlockType2 {
			return
		}
		if hb.Lines().Len() != 1 {
			return
		}
		firstLineSegment := hb.Lines().At(0)
		firstLine := firstLineSegment.Value(reader.Source())
		if !reIsTaskListComment.Match(firstLine) {
			return
		}
		if next := hb.NextSibling(); next == nil || next.Kind() != ast.KindList {
			return
		}
		// List -> Item -> TextBlock -> CheckBox
		list := hb.NextSibling().(*ast.List)
		recurse(list)
		return
	})
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (e *TaskList) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(extast.KindTaskCheckBox, e.renderTaskCheckBox)
}

// https://stackoverflow.com/a/28065534/3628322
func (e *TaskList) renderTaskCheckBox(
	w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*extast.TaskCheckBox)

	w.WriteString(`<input type=checkbox`)
	if n.IsChecked {
		w.WriteString(` checked`)
	}

	// List -> Item -> TextBlock -> CheckBox
	isTaskListItem := false
	if n.Parent() != nil && n.Parent().Parent() != nil {
		item := n.Parent().Parent().(*ast.ListItem)
		if any, ok := item.AttributeString(`class`); ok {
			if str, ok := any.(string); ok {
				isTaskListItem = strings.Contains(str, `task-list-item`) // not accurate
			}
		}
	}
	if isTaskListItem {
		n.SetAttributeString(`autocomplete`, `off`) // for firefox
	}
	// if !isTaskListItem {
	w.WriteString(` disabled`)
	// }

	if n.Attributes() != nil {
		html.RenderAttributes(w, n, nil)
	}
	_, _ = w.WriteString("> ")
	return ast.WalkContinue, nil
}

// func (e *TaskList) TransformHtml(doc *goquery.Document) error {
// 	doc.Find(`TaskList`).Each(func(i int, s *goquery.Selection) {
// 		replaced, err := e.single(s)
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		s.ReplaceWithSelection(replaced)
// 	})
// 	return nil
// }

// func (e *TaskList) single(s *goquery.Selection) (*goquery.Selection, error) {
// 	div := xhtml.Node{
// 		Type:     xhtml.ElementNode,
// 		DataAtom: atom.Div,
// 		Data:     `div`,
// 		Attr: []xhtml.Attribute{
// 			{
// 				Key: `class`,
// 				Val: `task-list`,
// 			},
// 		},
// 	}

// 	doc := goquery.NewDocumentFromNode(&div)
// 	doc.AppendSelection(s.Children())

// 	return doc.Selection, nil
// }
