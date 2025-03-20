package footnotes

import (
	"embed"
	"os"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = os.DirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		dynamic.WithStyles(`footnotes`, _embed, _root, `style.css`)
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
