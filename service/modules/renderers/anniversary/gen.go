package friends

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/theme/modules/sass"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css script.js
var _root embed.FS

func init() {
	dynamic.RegisterInit(func() {
		const module = `anniversary`
		dynamic.WithStyles(module, _root, `style.css`)
		dynamic.WithScripts(module, _root, `script.js`)
		sass.WatchDefaultAsync(string(dir.SourceAbsoluteDir()))
	})
}
