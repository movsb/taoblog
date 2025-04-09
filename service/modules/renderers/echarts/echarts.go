package echarts

//go:generate sh -c "test -s echarts.min.js || curl -LO https://cdn.jsdelivr.net/npm/echarts@5.6.0/dist/echarts.min.js"

import (
	"context"
	"embed"
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
	svg, err := render(context.Background(), string(source))
	if err != nil {
		return err
	}
	w.Write([]byte(svg))
	return nil
}

// https://echarts.apache.org/handbook/zh/how-to/cross-platform/server
func render(ctx context.Context, option string) (string, error) {
	script := `
(function() {
let option;
` + option + `
let chart = echarts.init(null,null, {
	renderer: 'svg',
	ssr: true,
	width: 800,
	height: 500,
});
chart.setOption(option);
return chart.renderToSVGString();
})();
`

	val, err := runtime().Execute(ctx, script)
	if err != nil {
		return ``, err
	}
	return val.Export().(string), nil
}
