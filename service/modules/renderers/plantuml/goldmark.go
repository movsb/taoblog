package plantuml

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

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
			if cb.Info != nil {
				info := string(cb.Info.Segment.Value(reader.Source()))
				if info == `plantuml` {
					puCodeBlocks = append(puCodeBlocks, cb)
				}
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

	light, dark, err := p.fetch(uml)
	if err != nil {
		p.error(writer)
		log.Println(`渲染失败`, err)
		return ast.WalkContinue, nil
	}

	writer.Write(light)
	writer.Write(dark)

	return ast.WalkContinue, nil
}

// TODO fallback 到用链接。
func (p *_PlantUMLRenderer) error(w util.BufWriter) {
	fmt.Fprintln(w, `<p style="color:red">PlantUML 渲染失败。</p>`)
}

func (p *_PlantUMLRenderer) fetch(source []byte) ([]byte, []byte, error) {
	compressed, err := compress(source)
	if err != nil {
		return nil, nil, err
	}

	var (
		content1, content2 []byte
		err1, err2         error
	)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	wg.Add(2)
	go func() {
		defer wg.Done()
		content1, err1 = fetch(ctx, p.server, p.format, compressed, false)
	}()
	go func() {
		defer wg.Done()
		content2, err2 = fetch(ctx, p.server, p.format, compressed, true)
	}()
	wg.Wait()

	// 全部错误才算错。
	if err1 != nil && err2 != nil {
		return nil, nil, errors.Join(err1, err2)
	}

	if len(content1) > 0 {
		content1 = style(content1, false)
	}
	if len(content2) > 0 {
		content2 = style(content2, true)
	}

	content1 = strip(content1)
	content2 = strip(content2)

	return content1, content2, nil
}
