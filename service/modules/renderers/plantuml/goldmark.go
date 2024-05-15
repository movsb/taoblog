package plantuml

import (
	"bytes"
	"context"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var _ interface {
	parser.ASTTransformer
	goldmark.Extender
	renderer.NodeRenderer
} = (*_PlantUMLRenderer)(nil)

type _PlantUMLRenderer struct {
	server string // 可以是 api 前缀
	format string
}

func New(server string, format string) *_PlantUMLRenderer {
	return &_PlantUMLRenderer{
		server: server,
		format: format,
	}
}

func (p *_PlantUMLRenderer) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(p, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(p, 100)))
}

type _PlantUMLRendererBlock struct {
	ast.BaseBlock
	ref *ast.FencedCodeBlock
}

var _plantUMLCodeBLockKind = ast.NewNodeKind(`plantuml_code_block`)

func (b *_PlantUMLRendererBlock) Kind() ast.NodeKind {
	return _plantUMLCodeBLockKind
}
func (b *_PlantUMLRendererBlock) Dump(source []byte, level int) {
	b.ref.Dump(source, level)
}

func (p *_PlantUMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(_plantUMLCodeBLockKind, p.renderCodeBlock)
}

// Transform implements parser.ASTTransformer.
func (p *_PlantUMLRenderer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	puCodeBlocks := []*ast.FencedCodeBlock{}
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindFencedCodeBlock {
			cb := n.(*ast.FencedCodeBlock)
			info := string(cb.Info.Segment.Value(reader.Source()))
			if info == `plantuml` {
				puCodeBlocks = append(puCodeBlocks, cb)
			}
		}
		return ast.WalkContinue, nil
	})
	for _, cb := range puCodeBlocks {
		cb.Parent().ReplaceChild(cb.Parent(), cb, &_PlantUMLRendererBlock{
			ref: cb,
		})
	}
}

// TODO 可选渲染成链接还是直接嵌入页面文件中，当前是后者。
func (p *_PlantUMLRenderer) renderCodeBlock(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n = n.(*_PlantUMLRendererBlock).ref
	b := bytes.NewBuffer(nil)
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		b.Write(line.Value(source))
	}
	uml := b.Bytes()
	if svg, err := p.asSvg(uml); err != nil {
		p.error(writer)
		log.Println(`渲染失败`)
		return ast.WalkContinue, nil
	} else {
		writer.Write(svg)
		return ast.WalkContinue, nil
	}
}

func (p *_PlantUMLRenderer) error(w util.BufWriter) {
	fmt.Fprintln(w, `<p style="color:red">PlantUML 渲染失败。</p>`)
}

func (p *_PlantUMLRenderer) asSvg(source []byte) ([]byte, error) {
	compress, err := compress(source)
	if err != nil {
		return nil, err
	}
	raw, err := fetch(context.Background(), p.server, p.format, compress)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
