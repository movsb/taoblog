package image_viewer

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
)

//go:embed vim.js
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `vim`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithScripts(module, `vim.js`)
	})
}
