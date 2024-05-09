package renderers

import (
	"github.com/yuin/goldmark/ast"
)

// 后面统一改成 Option。
type Option2 = any

type EnteringWalker interface {
	WalkEntering(n ast.Node) (ast.WalkStatus, error)
}

func appendClass(node ast.Node, name string) {
	any, found := node.AttributeString(name)
	if list, ok := any.(string); !found || ok {
		if list == "" {
			list = name
		} else {
			list += " "
			list += name
		}
		node.SetAttributeString(`class`, list)
	}
}

func Testing() Option {
	return func(me *_Markdown) error {
		me.testing = true
		return nil
	}
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
			appendClass(typed, `marker-`+class)
		}
	}
	return ast.WalkContinue, nil
}

// 保留列表样式。
//
// 只是增加类名，前端通过类名自行决定怎么展示。
func WithReserveListItemMarkerStyle() Option2 {
	return &_ReserveListItemMarkerStyle{}
}
