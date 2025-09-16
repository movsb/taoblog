package loaded

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
)

//go:embed script.js
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `loaded`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithScripts(module, `script.js`)
	})
}
