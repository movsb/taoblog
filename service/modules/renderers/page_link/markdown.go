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
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
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
		const module = `alerts`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
	})
}

type PageLink struct {
	ctx          context.Context
	getPageTitle GetPageTitle
}

type GetPageTitle func(ctx context.Context, id int) (string, error)

func New(ctx context.Context, getPageTitle GetPageTitle) *PageLink {
	return &PageLink{
		ctx:          ctx,
		getPageTitle: getPageTitle,
	}
}

var _ interface {
	goldmark.Extender
	parser.InlineParser
} = (*PageLink)(nil)

func (e *PageLink) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		// 优先级高于链接，低于代码。
		parser.WithInlineParsers(util.Prioritized(e, 150)),
	)
}

func (e *PageLink) Trigger() []byte {
	return []byte{'['}
}

var re = regexp.MustCompile(`\[\[(\d+)\]\]`)

func (e *PageLink) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	indices := re.FindSubmatch(line)
	if indices == nil {
		return nil
	}
	id, _ := strconv.Atoi(string(indices[1]))

	title, err := e.getPageTitle(e.ctx, id)
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

	return link
}
