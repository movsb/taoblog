package highlight

import (
	"embed"
	"fmt"
	"html"
	"sync"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
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
					// 因为 innerHTML 插入的 script 不会被执行，所以用这个手段。
					// 另外，鉴于 window.event 是被 deprecated 的，所以也不用。
					// https://developer.mozilla.org/en-US/docs/Web/API/Window/event
					// https://stackoverflow.com/q/12614862/3628322
					w.WriteString(fmt.Sprintf(
						`<img id="%[1]s" style="display:none;" src="https://" onerror="syncCodeScroll('%[1]s')"/>`,
						utils.RandomString(),
					))
					w.WriteRune('\n')
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
