package emojis

import (
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	//go:embed assets style.css
	_embed embed.FS

	_root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

	// 映射：狗头 → weixin/doge.png
	_refs = map[string]string{}
)

var BaseURLForDynamic = dynamic.BaseURL.JoinPath(`emojis/`)

type Emojis struct {
	baseURL *url.URL
}

var _ interface {
	goldmark.Extender
	parser.InlineParser
} = (*Emojis)(nil)

func New(baseURL *url.URL) *Emojis {
	return &Emojis{
		baseURL: baseURL,
	}
}

func init() {
	dynamic.RegisterInit(func() {
		const module = `emojis`

		assetsDirEmbed := utils.Must1(fs.Sub(_embed, `assets`))
		assetsDirRoot := utils.Must1(fs.Sub(_root, `assets`))

		dynamic.WithRoots(module, assetsDirEmbed, assetsDirRoot, _embed, _root)
		dynamic.WithStyles(module, `style.css`)

		weixin := func(fileName string, aliases ...string) {
			// NOTE：emoji 用的单数
			// NOTE：没有转码，图简单。
			dest := fmt.Sprintf(`weixin/%s`, fileName)
			for _, a := range aliases {
				_refs[a] = dest
			}
		}

		weixin(`doge.png`, `doge`, `旺柴`, `狗头`)
		weixin(`机智.png`, `机智`)
		weixin(`捂脸.png`, `捂脸`)
		weixin(`耶.png`, `耶`)
		weixin(`皱眉.png`, `皱眉`, `纠结`, `小纠结`)
	})
}

func (e *Emojis) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(e, 199),
	))
}

func (e *Emojis) Trigger() []byte {
	return []byte{'[', ']'}
}

var re = regexp.MustCompile(`(?U:^\[.+\])`)

func (e *Emojis) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	emoji := re.Find(line)
	if len(emoji) == 0 {
		return nil
	}
	link := ast.NewLink()
	link.Title = emoji[1 : len(emoji)-1]
	ref := _refs[string(link.Title)]
	if len(ref) == 0 {
		return nil
	}

	if u, err := url.Parse(ref); err == nil {
		u = e.baseURL.ResolveReference(u)
		link.Destination = []byte(u.String())
	}

	block.Advance(len(emoji))
	img := ast.NewImage(link)
	img.SetAttributeString(`class`, `emoji weixin`)
	return img
}

func (e *Emojis) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img.emoji.weixin`).Each(func(i int, s *goquery.Selection) {
		s.SetAttr(`alt`, fmt.Sprintf(`[%s]`, s.AttrOr(`title`, ``)))
	})
	return nil
}
