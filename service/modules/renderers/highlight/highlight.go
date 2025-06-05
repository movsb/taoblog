package highlight

import (
	"embed"
	"fmt"
	"html"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
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
	return backend()
}

var backend = sync.OnceValue(func() goldmark.Extender {
	return highlighting.NewHighlighting(
		// highlighting.WithCSSWriter(os.Stdout),
		highlighting.WithStyle(`onedark`),
		highlighting.WithFormatOptions(
			chromahtml.LineNumbersInTable(true),
			// 博客主题默认，不需要额外配置。
			// chromahtml.TabWidth(4),
			chromahtml.WithClasses(true),
			chromahtml.WithLineNumbers(true),
		),
		highlighting.WithWrapperRenderer(func(w util.BufWriter, context highlighting.CodeBlockContext, entering bool) {
			if entering {
				if context.Highlighted() {
					w.WriteString(`<div class="code-scroll-synchronizer">`)
					w.WriteString(gold_utils.InjectImage(`syncCodeScroll`))
				} else {
					language := string(utils.DropLast1(context.Language()))
					if language != "-" {
						w.WriteString(fmt.Sprintf(`<pre><code class="language-%s">`, html.EscapeString(language)))
					} else {
						w.WriteString(`<pre><code>`)
					}
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
