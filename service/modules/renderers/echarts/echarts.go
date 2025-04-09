package echarts

//go:generate sh -c "test -s echarts.min.js || curl -LO https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js"

import (
	"context"
	"embed"
	"fmt"
	"io"
	"sync"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/parser"
)

//go:embed echarts.min.js
var _embed embed.FS

var runtime = sync.OnceValue(func() *Runtime {
	rt, err := NewRuntime(context.Background(),
		utils.Must1(_embed.ReadFile(`echarts.min.js`)),
	)
	if err != nil {
		panic(err)
	}
	return rt
})

type _ECharts struct{}

func New() gold_utils.FencedCodeBlockRenderer {
	return &_ECharts{}
}

func (e *_ECharts) RenderFencedCodeBlock(w io.Writer, language string, attrs parser.Attributes, source []byte) error {
	var (
		width  = gold_utils.AttrIntOrDefault(attrs, `width`, 800)
		height = gold_utils.AttrIntOrDefault(attrs, `height`, 500)
	)

	svg, err := render(context.Background(), string(source), width, height)
	if err != nil {
		return err
	}
	w.Write([]byte(svg))
	return nil
}

// https://echarts.apache.org/handbook/zh/how-to/cross-platform/server
func render(ctx context.Context, option string, width, height int) (string, error) {
	script := fmt.Sprintf(`
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
