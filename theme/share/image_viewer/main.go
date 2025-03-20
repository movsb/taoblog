package image_viewer

import (
	"embed"
	"os"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed zoom-1.0.7.min.iife.js image-view.js style.css
var _embed embed.FS
var _root = os.DirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `image-viewer`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
		dynamic.WithScripts(module, `zoom-1.0.7.min.iife.js`, `image-view.js`)
	})
}
