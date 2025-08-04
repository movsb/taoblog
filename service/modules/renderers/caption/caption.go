package caption

import (
	"bytes"
	"embed"
	"log"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `caption`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

type _Caption struct {
	web gold_utils.WebFileSystem
}

func New(web gold_utils.WebFileSystem) *_Caption {
	return &_Caption{web: web}
}

func getSource(web gold_utils.WebFileSystem, path string) *proto.FileSpec_Meta_Source {
	f, err := web.OpenURL(path)
	if err != nil {
		log.Println(path, err)
		return nil
	}
	info := utils.Must1(f.Stat())
	sys, ok := info.Sys().(*models.File)
	if !ok {
		return nil
	}
	return sys.Meta.Source
}

func (m *_Caption) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img,video`).Each(func(i int, s *goquery.Selection) {
		if s.HasClass(`static`) {
			return
		}

		isParentP := s.Parent().Nodes[0].DataAtom == atom.P
		isOnlyChild := s.Parent().Children().Length() == 1
		if !(isParentP && isOnlyChild) {
			return
		}

		src := s.AttrOr(`src`, ``)
		if src == `` {
			return
		}

		url, err := url.Parse(src)
		if err != nil {
			log.Println(src, err)
			return
		}
		if url.EscapedPath() != url.String() {
			return
		}

		spec := getSource(m.web, src)
		if spec == nil {
			return
		}

		replace(s, spec)
	})
	return nil
}

var md = sync.OnceValue(func() goldmark.Markdown {
	return goldmark.New()
})

func replace(s *goquery.Selection, spec *proto.FileSpec_Meta_Source) {
	var (
		parent = s.Parent()
		obj    = s.Remove().Nodes[0]

		figure = &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Figure,
			Data:     `figure`,
		}
		caption = &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Figcaption,
			Data:     `figcaption`,
		}
	)

	figure.AppendChild(obj)
	figure.AppendChild(caption)
	parent.ReplaceWithNodes(figure)

	switch spec.Format {
	case proto.FileSpec_Meta_Source_Plaintext:
		cap := &html.Node{
			Type: html.TextNode,
			Data: spec.Caption,
		}
		caption.AppendChild(cap)
	case proto.FileSpec_Meta_Source_Markdown:
		buf := bytes.NewBuffer(nil)
		md().Convert([]byte(spec.Caption), buf)
		// log.Println(buf.String())
		ns, _ := html.ParseFragment(buf, caption)
		if len(ns) > 0 {
			caption.AppendChild(ns[0])
		}
	}
}
