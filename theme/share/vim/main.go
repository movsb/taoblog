package image_viewer

import (
	"embed"
	"os"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
)

//go:embed vim.js
var _embed embed.FS
var _root = os.DirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `vim`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithScripts(module, `vim.js`)
	})
}
