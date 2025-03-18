package goldmark_katex

import (
	"bufio"
	"bytes"
	"context"
	"time"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/phuslu/lru"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type Extender struct {
	r     *Runtime
	cache *Cache
}

func New(r *Runtime, c *Cache) goldmark.Extender {
	return &Extender{
		r:     r,
		cache: c,
	}
}

func (e *Extender) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(util.Prioritized(mathjax.NewMathJaxBlockParser(), 701)),
		parser.WithInlineParsers(util.Prioritized(mathjax.NewInlineMathParser(), 501)),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(util.Prioritized(&HTMLRenderer{r: e.r, b: &_Backend{}, c: e.cache}, 100)),
	)
}

type HTMLRenderer struct {
	r *Runtime
	b *_Backend
	c *Cache
}

func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(mathjax.KindInlineMath, func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		return ast.WalkSkipChildren, r.render(writer, source, n, entering, false)
	})
	reg.Register(mathjax.KindMathBlock, func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		return ast.WalkContinue, r.render(writer, source, n, entering, true)
	})
	mathjax.NewInlineMathRenderer(``, ``).RegisterFuncs(r.b)
	mathjax.NewMathBlockRenderer(``, ``).RegisterFuncs(r.b)
}

type _Backend struct {
	inline renderer.NodeRendererFunc
	block  renderer.NodeRendererFunc
}

func (r *_Backend) Register(k ast.NodeKind, f renderer.NodeRendererFunc) {
	switch k {
	case mathjax.KindInlineMath:
		r.inline = f
		return
	case mathjax.KindMathBlock:
		r.block = f
		return
	}
	panic(`unknown node kind`)
}

type CacheKey struct {
	DisplayMode bool
	Expression  string
}

type Cache = lru.TTLCache[CacheKey, []byte]

func (r *HTMLRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool, block bool) (outErr error) {
	if !entering {
		return nil
	}

	defer utils.CatchAsError(&outErr)

	wb := bytes.NewBuffer(nil)
	w2 := bufio.NewWriter(wb)
	utils.Must1(utils.IIF(block, r.b.block, r.b.inline)(w2, source, n, entering))
	w2.Flush()
	equation := wb.Bytes()
	if block {
		equation = bytes.TrimPrefix(equation, []byte(`<p><span class="math display">`))
	} else {
		equation = bytes.TrimPrefix(equation, []byte(`<span class="math inline">`))
	}

	html, err, _ := r.c.GetOrLoad(
		context.Background(),
		CacheKey{
			DisplayMode: block,
			Expression:  string(equation),
		},
		func(ctx context.Context, ck CacheKey) ([]byte, time.Duration, error) {
			html, err := Render(r.r, ck.Expression, ck.DisplayMode)
			return []byte(html), time.Hour, err
		},
	)
	if err != nil {
		return err
	}

	_, err = w.Write(html)
	return err
}

func Render(r *Runtime, equation string, displayMode bool) (_ string, outErr error) {
	defer utils.CatchAsError(&outErr)

	value := utils.Must1(r.Execute(
		context.Background(),
		`katex.renderToString(expression, { displayMode: displayMode, output: output })`,
		Argument{Name: `expression`, Value: equation},
		Argument{Name: `displayMode`, Value: displayMode},
		Argument{Name: `output`, Value: `mathml`},
	))

	return value.Export().(string), nil
}
