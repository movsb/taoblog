package emojis

import (
	"embed"
	"fmt"
	"regexp"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//go:embed assets style.css
var _root embed.FS

var (
	_once sync.Once
	_refs = map[string]string{}
)

type Emojis struct{}

var _ interface {
	goldmark.Extender
	parser.InlineParser
} = (*Emojis)(nil)

func New() *Emojis {
	_once.Do(initEmojis)
	return &Emojis{}
}

func initEmojis() {
	dynamic.Dynamic[`emojis`] = dynamic.Content{
		Root: _root,
		Styles: []string{
			string(utils.Must1(_root.ReadFile(`style.css`))),
		},
	}

	weixin := func(fileName string, aliases ...string) {
		// NOTE：emoji 用的单数
		// NOTE：没有转码，图简单。
		dest := fmt.Sprintf(`assets/weixin/%s`, fileName)
		for _, a := range aliases {
			_refs[a] = dest
		}
	}

	weixin(`doge.png`, `doge`, `旺柴`, `狗头`)
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
	link.Destination = []byte(`/v3/dynamic/` + ref)
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
