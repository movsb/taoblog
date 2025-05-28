package hashtags

import (
	"regexp"
	"sync"

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
		resolve: dropInvalid(resolve),
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
		if shouldDrop(tag) {
			continue
		}
		list = append(list, tag)
	}
	*t.out = list
}

var getRegexps = sync.OnceValue(func() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`^L\d+$`),
		regexp.MustCompile(`^L\d+-L\d+`),
	}
})

func shouldDrop(tag string) bool {
	for _, re := range getRegexps() {
		if re.MatchString(tag) {
			return true
		}
	}
	return false
}

// 丢弃部分可能无效的#️⃣标签。
//
// 像 [main.go#L123](https://...) 这种链接的文本中，是允许存在各种内联元素的，
// 包括但不限于 `code`, **加粗**，当然，#标签 也该是允许的。但是，把此处的 #L123 理解为
// 标签明显不太合理。
//
// 所以目前仅仅是简单地丢弃部分我觉得无效的标签，遇到了就补充。
func dropInvalid(resolver func(string) string) func(string) string {
	return func(tag string) string {
		if shouldDrop(tag) {
			return ``
		}
		return resolver(tag)
	}
}
