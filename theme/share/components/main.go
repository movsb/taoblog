package components

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
)

//go:embed geo-link.js
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `components`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithScripts(module, `geo-link.js`)
	})
}
