package friends

import (
	"embed"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/theme/modules/sass"
)

//go:generate sass --no-source-map style.scss style.css

//go:embed style.css script.js
var _root embed.FS

func init() {
	dynamic.RegisterInit(func() {
		dynamic.Dynamic[`anniversary`] = dynamic.Content{
			Styles: []string{
				string(utils.Must1(_root.ReadFile(`style.css`))),
			},
			Scripts: []string{
				string(utils.Must1(_root.ReadFile(`script.js`))),
			},
		}
		sass.WatchDefaultAsync(string(dir.SourceAbsoluteDir()))
	})
}
