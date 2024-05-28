package gold_utils

import (
	"slices"
	"strings"

	"github.com/yuin/goldmark/ast"
)

func AddClass(node ast.Node, classes ...string) {
	var class string
	if any, ok := node.AttributeString(`class`); ok {
		if str, ok := any.(string); ok {
			class = str
		}
	}
	classNames := strings.Fields(class)
	classNames = append(classNames, classes...)
	slices.Sort(classNames)
	classNames = slices.Compact(classNames)

	node.SetAttributeString(`class`, strings.Join(classNames, ` `))
}
