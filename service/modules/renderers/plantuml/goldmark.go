package plantuml

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	cache func(key string, loader func() (io.ReadCloser, error)) (io.ReadCloser, error)
}

func New(server string, format string, options ...Option) *_PlantUMLRenderer {
	p := &_PlantUMLRenderer{
		server: server,
		format: format,
	}
	for _, opt := range options {
		opt(p)
	}

	if p.cache == nil {
		p.cache = func(key string, loader func() (io.ReadCloser, error)) (io.ReadCloser, error) {
			return loader()
		}
	}

	return p
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

	compressed, err := compress(uml)
	if err != nil {
		p.error(writer)
		log.Println(`渲染失败`, err)
		return ast.WalkContinue, nil
	}

	got, err := p.cache(compressed, func() (io.ReadCloser, error) {
		light, dark, err := p.fetch(compressed)
		if err != nil {
			return nil, err
		}
		log.Println(`no using cache for plantuml ...`)
		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(_Cache{Light: light, Dark: dark}); err != nil {
			return nil, err
		}
		return io.NopCloser(buf), nil
	})
	if err != nil {
		p.error(writer)
		log.Println(`渲染失败`, err)
		return ast.WalkContinue, nil
	}

	defer got.Close()

	var cache _Cache
	if err := json.NewDecoder(got).Decode(&cache); err != nil {
		return ast.WalkStop, err
	}

	writer.Write(cache.Light)
	writer.Write(cache.Dark)

	return ast.WalkContinue, nil
}

// TODO fallback 到用链接。
func (p *_PlantUMLRenderer) error(w util.BufWriter) {
	fmt.Fprintln(w, `<p style="color:red">PlantUML 渲染失败。</p>`)
}

func (p *_PlantUMLRenderer) fetch(compressed string) ([]byte, []byte, error) {
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
