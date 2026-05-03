package highlight

import (
	"embed"
	"fmt"
	"html"
	"slices"
	"strings"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//go:generate sass --no-source-map --style compressed style.scss style.css

//go:embed style.css script.js
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `highlight`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
		dynamic.WithScripts(module, `script.js`)
	})
}

func New() goldmark.Extender {
	return _Extender{}
}

type _Extender struct{}

func (e _Extender) Extend(m goldmark.Markdown) {
	backend().Extend(m)
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(e, 999)))
}

// 通过统计行数来确定gutter的宽度。
// 下面的 wrapper 拿不到原始代码，只得在这里统计了。
func (e _Extender) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindFencedCodeBlock {
			lines := n.Lines().Len()
			if lines > 99 {
				gold_utils.AddClass(n, `gutter-3`)
			}
		}
		return ast.WalkContinue, nil
	})
}

var backend = sync.OnceValue(func() goldmark.Extender {
	return highlighting.NewHighlighting(
		// highlighting.WithCSSWriter(os.Stdout),
		highlighting.WithStyle(`onedark`),
		highlighting.WithFormatOptions(
			chromahtml.LineNumbersInTable(true),
			// 主题默认，不需要额外配置。
			// chromahtml.TabWidth(4),
			chromahtml.WithClasses(true),
			chromahtml.WithLineNumbers(true),
		),
		highlighting.WithWrapperRenderer(func(w util.BufWriter, context highlighting.CodeBlockContext, entering bool) {
			if entering {
				var hasGutter3 bool
				if attrs := context.Attributes(); attrs != nil {
					anyClass, _ := attrs.GetString(`class`)
					strClass, _ := anyClass.(string)
					hasGutter3 = slices.Contains(strings.Fields(strClass), `gutter-3`)
				}

				if context.Highlighted() {
					w.WriteString(`<div`)

					classes := []string{`code-scroll-synchronizer`}
					if hasGutter3 {
						classes = append(classes, `gutter-3`)
					}
					fmt.Fprintf(w, ` class="%s"`, strings.Join(classes, ` `))

					w.WriteString(`>`)
					w.WriteString(gold_utils.InjectImage(`syncCodeScroll`))
				} else {
					w.WriteString(`<pre><code`)

					classes := []string{}
					language := string(utils.DropLast1(context.Language()))
					if language != "-" && language != "" {
						classes = append(classes, fmt.Sprintf(`language-%s`, html.EscapeString(language)))
					}
					if hasGutter3 {
						classes = append(classes, `gutter-3`)
					}
					if len(classes) > 0 {
						fmt.Fprintf(w, ` class="%s"`, strings.Join(classes, ` `))
					}

					w.WriteString(`>`)
				}
			} else {
				if context.Highlighted() {
					w.WriteString(`</div>`)
					w.WriteRune('\n')
				} else {
					w.WriteString(`</code></pre>`)
				}
			}
		}),
	)
})
