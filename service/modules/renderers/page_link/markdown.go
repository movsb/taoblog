package page_link

import (
	"context"
	"embed"
	"fmt"
	"html"
	"log"
	"regexp"
	"strconv"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//go:embed style.css
var _embed embed.FS
var _local = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `page-link`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
	})
}

type PageLink struct {
	ctx          context.Context
	getPageTitle GetPageTitle
	refs         *[]int32

	selfAddedLinks map[*ast.Link]struct{}
}

type GetPageTitle = func(ctx context.Context, id int32) (string, error)

// ctx 用于 GetPageTitle
// refs: 用于输出引用。未排序、未去重。
func New(ctx context.Context, getPageTitle GetPageTitle, refs *[]int32) *PageLink {
	return &PageLink{
		ctx:          ctx,
		getPageTitle: getPageTitle,
		refs:         refs,

		selfAddedLinks: map[*ast.Link]struct{}{},
	}
}

var _ interface {
	goldmark.Extender
	parser.InlineParser
} = (*PageLink)(nil)

func (e *PageLink) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		// 优先级高于链接(200)，低于代码(100)。
		parser.WithInlineParsers(util.Prioritized(e, 150)),
		parser.WithASTTransformers(util.Prioritized(e, 999)),
	)
}

func (e *PageLink) Trigger() []byte {
	return []byte{'['}
}

var re = regexp.MustCompile(`^\[\[(\d+)\]\]`)

func (e *PageLink) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	indices := re.FindSubmatch(line)
	if indices == nil {
		return nil
	}

	id, _ := strconv.Atoi(string(indices[1]))

	title, err := e.getPageTitle(e.ctx, int32(id))
	if err != nil {
		log.Println(`获取文章标题失败：`, err)
		title = `页面不存在`
	}

	// TODO parse 的时候就做 render 的事，不好！
	link := ast.NewLink()
	// link.Title = []byte(title)
	link.Destination = []byte(fmt.Sprintf(`/%d/`, id))
	child := ast.NewString([]byte(html.EscapeString(title)))
	child.SetCode(true)
	link.AppendChild(link, child)
	// if err != nil {
	// 	link.SetAttributeString(`class`, `error`)
	// }

	block.Advance(len(indices[0]))

	if e.refs != nil {
		*e.refs = append(*e.refs, int32(id))
		e.selfAddedLinks[link] = struct{}{}
	}

	return link
}

var isID = regexp.MustCompile(`^/(\d+)/?$`)

// 同时处理形如 `[文本](/编号/)` 形式的内部引用。
// ApplyHtmlTransformers 是渲染后才得到的，如果 noRendering，则可能不会执行，
// 所以这里用 ASTTransformer。
func (e *PageLink) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if e.refs == nil {
		return
	}
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindLink {
			return ast.WalkContinue, nil
		}
		link := n.(*ast.Link)
		if _, ok := e.selfAddedLinks[link]; ok {
			return ast.WalkContinue, nil
		}
		// log.Println(string(link.Destination))
		match := isID.FindStringSubmatch(string(link.Destination))
		if match != nil {
			id := utils.Must1(strconv.Atoi(match[1]))
			*e.refs = append(*e.refs, int32(id))
		}
		return ast.WalkContinue, nil
	})
}
