package custom_break

import (
	"html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A CustomBreak struct represents a thematic break of Markdown text.
type CustomBreak struct {
	ast.BaseBlock

	content string
}

// Dump implements Node.Dump .
func (n *CustomBreak) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// KindCustomBreak is a NodeKind of the CustomBreak node.
var KindCustomBreak = ast.NewNodeKind("CustomBreak")

// Kind implements Node.Kind.
func (n *CustomBreak) Kind() ast.NodeKind {
	return KindCustomBreak
}

// NewCustomBreak returns a new CustomBreak node.
func NewCustomBreak(content string) *CustomBreak {
	return &CustomBreak{
		BaseBlock: ast.BaseBlock{},
		content:   content,
	}
}

// ----------

type CustomBreakParser struct {
}

var defaultCustomBreakParser = &CustomBreakParser{}

// NewCustomBreakParser returns a new BlockParser that
// parses thematic breaks.
func NewCustomBreakParser() parser.BlockParser {
	return defaultCustomBreakParser
}

var re = regexp.MustCompile(`^\s{0,3}---\s*(.+)\s*---\s*$`)

func isCustomBreak(line []byte, offset int) (string, bool) {
	w, pos := util.IndentWidth(line, offset)
	if w > 3 {
		return "", false
	}
	if matches := re.FindSubmatch(line[pos:]); matches != nil {
		return string(matches[1]), true
	}
	return "", false
}

func (b *CustomBreakParser) Trigger() []byte {
	return []byte{'-'}
}

func (b *CustomBreakParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	if content, ok := isCustomBreak(line, reader.LineOffset()); ok {
		reader.Advance(segment.Len() - 1)
		return NewCustomBreak(strings.TrimSpace(content)), parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (b *CustomBreakParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (b *CustomBreakParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

func (b *CustomBreakParser) CanInterruptParagraph() bool {
	return false
}

func (b *CustomBreakParser) CanAcceptIndentedLine() bool {
	return false
}

// ----

type _CustomBreak struct{}

var _ interface {
	goldmark.Extender
	renderer.NodeRenderer
} = (*_CustomBreak)(nil)

// 自定义分割线。
//
// 格式：
//
// --- 内容 ---
//
// TODO：改成 Paragraph Transformer。
func New() *_CustomBreak {
	return &_CustomBreak{}
}

func (b *_CustomBreak) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(defaultCustomBreakParser, 100)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(b, 100),
		),
	)
}

func (b *_CustomBreak) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCustomBreak, renderCustomBreak)
}

func renderCustomBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	br := n.(*CustomBreak)

	w.WriteString(`<div class="divider"><span>`)
	w.WriteString(html.EscapeString(br.content))
	w.WriteString(`</span></div>`)

	return ast.WalkContinue, nil
}
