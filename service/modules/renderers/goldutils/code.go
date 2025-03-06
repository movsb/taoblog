package gold_utils

import (
	"bytes"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type FencedCodeBlockRenderer interface {
	RenderFencedCodeBlock(w io.Writer, language string, attrs parser.Attributes, source []byte) error
}

////////////////////////////////////////////////////////////////////////////////

var _replacedFencedCodeBlockKind = ast.NewNodeKind(`replaced_fenced_code_block`)

type _ReplacedFencedCodeBlock struct {
	ast.BaseBlock
	ref      *ast.FencedCodeBlock
	r        FencedCodeBlockRenderer
	language string
	attrs    parser.Attributes
}

func (b *_ReplacedFencedCodeBlock) Kind() ast.NodeKind {
	return _replacedFencedCodeBlockKind
}
func (b *_ReplacedFencedCodeBlock) Dump(source []byte, level int) {
	b.ref.Dump(source, level)
}

////////////////////////////////////////////////////////////////////////////////

type FencedCodeBlockExtender struct {
	Renders *map[string]FencedCodeBlockRenderer
}

var _ interface {
	parser.ASTTransformer
	goldmark.Extender
	renderer.NodeRenderer
} = (*FencedCodeBlockExtender)(nil)

func (e *FencedCodeBlockExtender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(e, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(e, 100)))
}

func (e *FencedCodeBlockExtender) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(_replacedFencedCodeBlockKind, e.renderCodeBlock)
}

func (e *FencedCodeBlockExtender) renderCodeBlock(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	nn := n.(*_ReplacedFencedCodeBlock)
	buf := bytes.NewBuffer(nil)
	for i := range nn.ref.Lines().Len() {
		line := nn.ref.Lines().At(i)
		buf.Write(line.Value(source))
	}

	if err := nn.r.RenderFencedCodeBlock(writer, nn.language, nn.attrs, buf.Bytes()); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

// Transform implements parser.ASTTransformer.
func (e *FencedCodeBlockExtender) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	rCodeBlocks := []*_ReplacedFencedCodeBlock{}
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindFencedCodeBlock {
			cb := n.(*ast.FencedCodeBlock)
			if cb.Info != nil {
				info := cb.Info.Segment.Value(reader.Source())
				language, unparsed, _ := bytes.Cut(info, []byte{' '})
				if r, ok := (*e.Renders)[string(language)]; ok {
					attrs, _ := parser.ParseAttributes(text.NewReader(unparsed))
					rCodeBlocks = append(rCodeBlocks, &_ReplacedFencedCodeBlock{
						ref:      cb,
						r:        r,
						language: string(language),
						attrs:    attrs,
					})
				}
			}
		}
		return ast.WalkContinue, nil
	})
	for _, cb := range rCodeBlocks {
		cb.ref.Parent().ReplaceChild(cb.ref.Parent(), cb.ref, cb)
	}
}
