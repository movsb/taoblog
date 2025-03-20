package hashtags

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/hashtag"
)

// 自动解析 HashTags。
// tags: 包含 #。
// resolve：把标签解析成链接。
func New(resolve func(tag string) string, tags *[]string) any {
	return &_HashTags{
		resolve: resolve,
		out:     tags,
	}
}

var _ interface {
	parser.ASTTransformer
} = (*_HashTags)(nil)

type _HashTags struct {
	resolve func(tag string) string
	out     *[]string
}

func (t *_HashTags) Extend(m goldmark.Markdown) {
	(&hashtag.Extender{Resolver: t}).Extend(m)
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(t, 999)))
}

func (t *_HashTags) ResolveHashtag(n *hashtag.Node) ([]byte, error) {
	return []byte(t.resolve(string(n.Tag))), nil
}

func (t *_HashTags) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if t.out == nil {
		return
	}

	tags := map[string]struct{}{}
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == hashtag.Kind {
			tags[string(n.(*hashtag.Node).Tag)] = struct{}{}
		}
		return ast.WalkContinue, nil
	})
	list := make([]string, 0, len(tags))
	for tag := range tags {
		list = append(list, tag)
	}
	*t.out = list
}
