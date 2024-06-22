package emojis

import (
	"embed"
	"fmt"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
	"github.com/yuin/goldmark/parser"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//go:embed assets style.css
var _root embed.FS

var (
	_once sync.Once
	_refs []parser.Reference
)

type Emojis struct{}

var _ interface {
	renderers.ContextPreparer
	gold_utils.HtmlTransformer
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
		dest := fmt.Sprintf(`emoji:///assets/weixin/%s`, fileName)
		for _, a := range aliases {
			ref := parser.NewReference([]byte(a), []byte(dest), []byte(a))
			_refs = append(_refs, ref)
		}
	}

	weixin(`doge.png`, `doge`, `旺柴`, `狗头`)
}

func (e *Emojis) PrepareContext(ctx parser.Context) {
	for _, r := range _refs {
		ctx.AddReference(r)
	}
}

func (e *Emojis) TransformHtml(doc *goquery.Document) error {
	transform := func(a *goquery.Selection, u *url.URL) {
		img := goquery.NewDocumentFromNode(&html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Img,
			Data:     `img`,
		})
		img.SetAttr(`class`, `emoji weixin`)
		img.SetAttr(`title`, a.AttrOr(`title`, ``))
		u.Scheme = ""
		img.SetAttr(`src`, `/v3/dynamic`+u.String())
		a.ReplaceWithSelection(img.Selection)
	}
	doc.Find(`a`).Each(func(i int, s *goquery.Selection) {
		u, err := url.Parse(s.AttrOr(`href`, ``))
		if err != nil {
			return
		}
		if u.Scheme != `emoji` || u.Path == "" {
			return
		}
		transform(s, u)
	})
	return nil
}
