package footnotes

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `footnotes`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

type Extender struct {
	goldmark.Extender
}

func New() goldmark.Extender {
	return &Extender{
		Extender: extension.NewFootnote(
			extension.WithFootnoteBacklinkHTML(`^`),
			// NOTE：在同一个 HTML 页面中显示多篇文章的时候需要区别此。
			// NOTE：此时不显示更好。
			// extension.WithFootnoteIDPrefix(fmt.Sprintf(`article-%d-`, postId)),
		),
	}
}
