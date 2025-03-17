package friends

import (
	"embed"
	"os"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css script.js
var _embed embed.FS
var _root = os.DirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `anniversary`
		dynamic.WithStyles(module, _embed, _root, `style.css`)
		dynamic.WithScripts(module, _embed, _root, `script.js`)
	})
}
