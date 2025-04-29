package graph_viz

import (
	"bytes"
	"context"
	"embed"
	"io"

	"github.com/goccy/go-graphviz"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
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

func (gg *GraphViz) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) (outErr error) {
	defer utils.CatchAsError(&outErr)

	gv := utils.Must1(graphviz.New(context.Background()))
	defer gv.Close()

	graph := utils.Must1(graphviz.ParseBytes(source))
	graph.SetPad(.2)

	buf := bytes.NewBuffer(nil)
	buf.WriteString(`<div class="graphviz">`)
	utils.Must(gv.Render(context.Background(), graph, graphviz.SVG, buf))
	buf.WriteString(`</div>`)

	_, err := w.Write(buf.Bytes())
	return err
}
