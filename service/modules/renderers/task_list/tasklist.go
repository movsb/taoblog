package task_list

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
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

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css script.js
var _embed embed.FS
var _root = os.DirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `task-list`
		dynamic.WithStyles(module, _embed, _root, `style.css`)
		dynamic.WithScripts(module, _embed, _root, `script.js`)
	})
}

// 任务/待办列表（TODO list）。
//
// 支持标记任务在原文中的位置，以方便在页面上点击后直接修改任务状态。
//
// 格式：
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

// var reIsTaskListComment = regexp.MustCompile(`(?i:^\s{0,3}<!--\s*(?:TaskList|TODO)\s*-->\s*$)`)

// ... implements parser.ASTTransformer.
// TODO 只处理第一层任务。
func (e *TaskList) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	recurse := func(list *ast.List, preserveSource bool) {
		hasTaskItem := false
		for c := list.FirstChild(); c != nil; c = c.NextSibling() {
			item, ok := c.(*ast.ListItem)
			if !ok {
				continue
			}
			if item.FirstChild() == nil || item.FirstChild().Kind() != ast.KindTextBlock && item.FirstChild().Kind() != ast.KindParagraph {
				continue
			}
			first := item.FirstChild()
			if first.FirstChild() == nil || first.FirstChild().Kind() != extast.KindTaskCheckBox {
				continue
			}
			checkBox := first.FirstChild().(*extast.TaskCheckBox)
			classes := []string{`task-list-item`}
			if checkBox.IsChecked {
				classes = append(classes, `checked`)
			}
			gold_utils.AddClass(item, classes...)
			if preserveSource {
				item.SetAttributeString(`data-source-position`, fmt.Sprint(checkBox.Segment.Start))
			}
			hasTaskItem = true
		}
		if hasTaskItem {
			gold_utils.AddClass(list, `task-list`)
			if preserveSource {
				gold_utils.AddClass(list, `with-source-positions`)
			}
		}
	}

	ast.Walk(node, func(n ast.Node, entering bool) (status ast.WalkStatus, err error) {
		status = ast.WalkContinue
		if !entering {
			return
		}

		if n.Kind() != ast.KindList {
			return
		}

		// prevIsComment := func(n ast.Node) (yes bool) {
		// 	if n == nil {
		// 		return
		// 	}
		// 	if n.Kind() != ast.KindHTMLBlock {
		// 		return
		// 	}
		// 	hb := n.(*ast.HTMLBlock)
		// 	// https://spec.commonmark.org/0.31.2/#html-block
		// 	if hb.HTMLBlockType != ast.HTMLBlockType2 {
		// 		return
		// 	}
		// 	if hb.Lines().Len() != 1 {
		// 		return
		// 	}
		// 	firstLineSegment := hb.Lines().At(0)
		// 	firstLine := firstLineSegment.Value(reader.Source())
		// 	if !reIsTaskListComment.Match(firstLine) {
		// 		return
		// 	}
		// 	return true
		// }(n.PreviousSibling())

		// List -> Item -> TextBlock -> CheckBox
		// List -> Item -> Paragraph -> CheckBox
		recurse(n.(*ast.List), true)
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

	// List -> Item -> TextBlock/Paragraph -> CheckBox
	isTaskList := false
	if n.Parent() != nil && n.Parent().Parent() != nil && n.Parent().Parent().Parent() != nil {
		list := n.Parent().Parent().Parent().(*ast.List)
		if any, ok := list.AttributeString(`class`); ok {
			if str, ok := any.(string); ok {
				isTaskList = strings.Contains(str, `with-source-positions`) // not accurate
			}
		}
	}
	if isTaskList {
		n.SetAttributeString(`autocomplete`, `off`) // for firefox
	}
	// footer 里面有脚本，防止在预览文章的时候没加载脚本也可点击完成任务。
	// if !isTaskListItem {
	w.WriteString(` disabled`)
	// }

	if n.Attributes() != nil {
		html.RenderAttributes(w, n, nil)
	}
	_, _ = w.WriteString("> ")
	return ast.WalkContinue, nil
}
