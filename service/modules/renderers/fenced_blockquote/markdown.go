package fenced_blockquote

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Extender struct{}

func New() *Extender {
	return &Extender{}
}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(NewParser(), 999)),
	)
}

type Parser struct{}

func NewParser() parser.BlockParser {
	return &Parser{}
}

type fenceData struct {
	node   ast.Node
	length int
}

var fenceDataKey = parser.NewContextKey()

func (p *Parser) Trigger() []byte {
	return []byte{'"'}
}

func (p *Parser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	length, ok := matchOpeningFence(line, reader.LineOffset())
	if !ok {
		return nil, parser.NoChildren
	}
	advanceFenceLine(reader, segment, line)
	node := ast.NewBlockquote()
	pushFenceData(pc, fenceData{node: node, length: length})
	return node, parser.HasChildren
}

func (p *Parser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	data, ok := currentFenceData(pc, node)
	if !ok {
		return parser.Close
	}
	line, segment := reader.PeekLine()
	if matchClosingFence(line, reader.LineOffset(), data.length) {
		advanceFenceLine(reader, segment, line)
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

func (p *Parser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	popFenceData(pc, node)
}

func (p *Parser) CanInterruptParagraph() bool {
	return true
}

func (p *Parser) CanAcceptIndentedLine() bool {
	return false
}

func matchOpeningFence(line []byte, offset int) (int, bool) {
	w, pos := util.IndentWidth(line, offset)
	if w > 3 || pos >= len(line) || line[pos] != '"' {
		return 0, false
	}
	i := pos
	for i < len(line) && line[i] == '"' {
		i++
	}
	length := i - pos
	if length < 3 {
		return 0, false
	}
	if !util.IsBlank(line[i:]) {
		return 0, false
	}
	return length, true
}

func matchClosingFence(line []byte, offset, want int) bool {
	length, ok := matchOpeningFence(line, offset)
	return ok && length == want
}

func advanceFenceLine(reader text.Reader, segment text.Segment, line []byte) {
	newline := 1
	if len(line) == 0 || line[len(line)-1] != '\n' {
		newline = 0
	}
	reader.Advance(segment.Stop - segment.Start - newline + segment.Padding)
}

func pushFenceData(pc parser.Context, data fenceData) {
	stack, _ := pc.Get(fenceDataKey).([]fenceData)
	pc.Set(fenceDataKey, append(stack, data))
}

func currentFenceData(pc parser.Context, node ast.Node) (fenceData, bool) {
	stack, _ := pc.Get(fenceDataKey).([]fenceData)
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].node == node {
			return stack[i], true
		}
	}
	return fenceData{}, false
}

func popFenceData(pc parser.Context, node ast.Node) {
	stack, _ := pc.Get(fenceDataKey).([]fenceData)
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].node == node {
			stack = append(stack[:i], stack[i+1:]...)
			pc.Set(fenceDataKey, stack)
			return
		}
	}
}
