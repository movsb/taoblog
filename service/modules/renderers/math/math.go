package katex

import (
	"context"
	"embed"
	"io/fs"
	"sync"
	"time"

	gold_katex "github.com/libkush/goldmark-katex"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/theme/modules/sass"
	"github.com/phuslu/lru"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

//go:generate sass --style compressed --no-source-map katex/style.scss katex/style.css

//go:embed static katex/katex.min.stripped.css katex/style.css
var Root embed.FS

func init() {
	dynamic.RegisterInit(func() {
		const module = `katex`
		katexDir := utils.Must1(fs.Sub(Root, `katex`))
		dynamic.WithRoot(module, utils.Must1(fs.Sub(Root, `static`)))
		dynamic.WithStyles(module, katexDir, `katex.min.stripped.css`, `style.css`)
		sass.WatchDefaultAsync(dir.SourceAbsoluteDir().Join(`katex`))
	})
}

type Math struct{}

func New() goldmark.Extender {
	return &Math{}
}

func (e *Math) Extend(m goldmark.Markdown) {
	_onceCache.Do(func() {
		_cache = lru.NewTTLCache[_CacheKey, string](1024)
	})

	exec := gold_katex.New_Exec()
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&gold_katex.Parser{}, 0),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&HTMLRenderer{
				exec: exec,
			}, 0),
		),
	)
}

type _CacheKey struct {
	displayMode bool
	tex         string
}

var (
	_cache     *lru.TTLCache[_CacheKey, string]
	_onceCache sync.Once
)

// 优化：
//
//   - 共用了代码
//   - 简化了缓存
//   - 默认输出 HTML 而不是 mathml
type HTMLRenderer struct {
	exec *gold_katex.Exec
}

func (r *HTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(gold_katex.KindInline, func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		return ast.WalkContinue, r.render(writer, source, n, entering, false)
	})
	reg.Register(gold_katex.KindBlock, func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
		return ast.WalkContinue, r.render(writer, source, n, entering, true)
	})
}

func (r *HTMLRenderer) render(w util.BufWriter, _ []byte, n ast.Node, entering bool, block bool) error {
	if !entering {
		return nil
	}

	var equation string

	if block {
		equation = string(n.(*gold_katex.Block).Equation)
	} else {
		equation = string(n.(*gold_katex.Inline).Equation)
	}

	html, err, _ := _cache.GetOrLoad(
		context.Background(),
		_CacheKey{
			displayMode: block,
			tex:         string(equation),
		},
		func(ctx context.Context, ck _CacheKey) (string, time.Duration, error) {
			html, err := Render(equation, block, r.exec)
			return html, time.Hour, err
		},
	)
	if err != nil {
		return err
	}

	_, err = w.WriteString(html)
	return err
}

func Render(equation string, displayMode bool, exec *gold_katex.Exec) (string, error) {
	res, err := exec.RunString(
		`katex.renderToString(expression, { displayMode: displayMode, output: 'html' })`,
		gold_katex.Arg{Name: `expression`, Value: equation},
		gold_katex.Arg{Name: `displayMode`, Value: displayMode},
	)
	if err != nil {
		return ``, err
	}
	return res.Export().(string), nil
}
