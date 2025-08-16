package graph_viz

import (
	"bytes"
	"context"
	"embed"
	"io"
	"sync"

	"github.com/goccy/go-graphviz"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/parser"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `graphviz`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

type GraphViz struct{}

func New() *GraphViz {
	return &GraphViz{}
}

// TODO close global instance
var gv = sync.OnceValue(func() *graphviz.Graphviz {
	return utils.Must1(graphviz.New(context.Background()))
})

func (gg *GraphViz) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) error {
	graph, err := graphviz.ParseBytes(source)
	if err != nil {
		gold_utils.RenderError(w, err)
		return nil
	}

	graph.SetPad(.2)

	buf := bytes.NewBuffer(nil)
	buf.WriteString(`<div class="graphviz">`)
	if err := gv().Render(context.Background(), graph, graphviz.SVG, buf); err != nil {
		gold_utils.RenderError(w, err)
		return nil
	}
	buf.WriteString(`</div>`)

	_, err = w.Write(buf.Bytes())
	return err
}
