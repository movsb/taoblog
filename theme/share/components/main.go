package components

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed geo-link.js datetime-picker.js style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `components`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithScripts(module, `geo-link.js`, `datetime-picker.js`)
		dynamic.WithStyles(module, `style.css`)
	})
}
