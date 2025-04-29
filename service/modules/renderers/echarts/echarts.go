package echarts

//go:generate sh -c "test -s echarts.min.js || curl -LO https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js"

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/parser"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `echarts`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

//go:embed echarts.min.js style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

type _ECharts struct {
	cache cache.Getter
}

func New(cache cache.Getter) gold_utils.FencedCodeBlockRenderer {
	return &_ECharts{
		cache: cache,
	}
}

type CacheKey struct {
	Source string
}

type _CacheValue struct {
	SVG string
}

func (e *_ECharts) RenderFencedCodeBlock(w io.Writer, language string, attrs parser.Attributes, source []byte) error {
	var (
		width  = gold_utils.AttrIntOrDefault(attrs, `width`, 800)
		height = gold_utils.AttrIntOrDefault(attrs, `height`, 500)
	)

	var value _CacheValue
	if err := e.cache(
		CacheKey{string(source)}, time.Hour*24*7, &value,
		func() (any, error) {
			log.Println(`未使用缓存渲染 ECharts`)
			svg, err := render(context.Background(), string(source), width, height)
			return _CacheValue{SVG: svg}, err
		}); err != nil {
		return err
	}

	if len(value.SVG) < 4 {
		return nil
	}

	// <svg ... ➡️ <svg class="echarts"
	w.Write([]byte(`<svg class="echarts"`))
	w.Write([]byte(value.SVG[len(`<svg`):]))
	return nil
}

// https://echarts.apache.org/handbook/zh/how-to/cross-platform/server
func render(ctx context.Context, option string, width, height int) (string, error) {
	script := fmt.Sprintf(`
// 从 ECharts 官网拷过来的示例代码没有 let 定义，为了避免重复定义或清空
// 后图表没有重置，这里强制置空。
(function() { option={}; })();
(function() {
%s
let chart = echarts.init(null,null, {
	renderer: 'svg',
	ssr: true,
	width: %d,
	height: %d,
});
typeof option != 'undefined' && chart.setOption(option);
return chart.renderToSVGString();
})();
`, option, width, height)

	val, err := runtime().Execute(ctx, script)
	if err != nil {
		return ``, err
	}

	return val.Export().(string), nil
}

var runtime = sync.OnceValue(func() *Runtime {
	rt, err := NewRuntime(context.Background(),
		utils.Must1(_embed.ReadFile(`echarts.min.js`)),
	)
	if err != nil {
		panic(err)
	}
	return rt
})
