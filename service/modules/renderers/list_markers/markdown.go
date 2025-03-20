package list_markers

import (
	"embed"
	"os"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/ast"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = os.DirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `list-markers`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(`list-markers`, `style.css`)
	})
}

type _ReserveListItemMarkerStyle struct{}

var knownListItemMarkers = map[byte]string{
	'-': `minus`,
	'+': `plus`,
	'*': `asterisk`,
	'.': `period`,
	')': `parenthesis`,
}

func (*_ReserveListItemMarkerStyle) WalkEntering(n ast.Node) (ast.WalkStatus, error) {
	switch typed := n.(type) {
	case *ast.List:
		if class, ok := knownListItemMarkers[typed.Marker]; ok {
			gold_utils.AddClass(typed, `marker-`+class)
		}
	}
	return ast.WalkContinue, nil
}

// 保留列表样式。
//
// 只是增加类名，前端通过类名自行决定怎么展示。
func New() *_ReserveListItemMarkerStyle {
	return &_ReserveListItemMarkerStyle{}
}
