package media_tags

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dhowden/tag"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"golang.org/x/net/html/atom"
)

var SourceRelativeDir = dir.SourceRelativeDir()

//go:embed player.html script.js style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `media_tags`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
		dynamic.WithScripts(module, `script.js`)
	})
}

type MediaTags struct {
	web  gold_utils.WebFileSystem
	tmpl *utils.TemplateLoader
}

var _gTmpl *utils.TemplateLoader

type Option func(*MediaTags)

var t = sync.OnceValue(func() *utils.TemplateLoader {
	return utils.NewTemplateLoader(utils.IIF(version.DevMode(), _root, fs.FS(_embed)), nil, dynamic.Reload)
})

func New(web gold_utils.WebFileSystem, options ...Option) *MediaTags {
	tag := &MediaTags{
		web: web,
	}

	for _, opt := range options {
		opt(tag)
	}

	// 判断为空的目的是测试里面可能会预初始化。
	if _gTmpl == nil {
		_gTmpl = t()
	}

	return &MediaTags{
		web:  web,
		tmpl: _gTmpl,
	}
}

func (t *MediaTags) TransformHtml(doc *goquery.Document) error {
	var outErr error
	doc.Find(`audio`).EachWithBreak(func(_ int, s *goquery.Selection) bool {
		src := t.getSrc(s)
		if src == "" {
			log.Println(`没有找到资源。`)
			return true
		}
		md, err := t.parse(src)
		if err != nil {
			// 忽略错误，继续
			log.Println("解析数据出错：", err)
			return true
		}

		// 类似下面这样的 audio 是 block 级别元素。
		//
		// <audio>
		//   <source>
		// </audio>
		//
		// 而类似下面却是 inline 元素。
		//
		// <audio></audio>
		// 或
		// AAA <audio></audio> ZZZ
		//
		// 所以目前处理所以类型的 audio，而不仅是 block 级别。
		// 仅当是唯一子元素（即 block）时，才渲染。
		//
		// https://spec.commonmark.org/dingus/?text=123%0A%0A%3Caudio%20controls%3E%3C%2Faudio%3E%0A%0A%3Caudio%20controls%3E%0A%3C%2Faudio%3E%0A%0Aasdf
		//
		// 如果父元素是 p 且只有一个子元素且是 audio，则认为是 block，
		// 或者：html.Parse 会默认 parse 到 body 下，如果父元素是 body，则也认为是 block。
		isParentBody := s.Parent().Nodes[0].DataAtom == atom.Body
		isOnlyChild := s.Parent().Nodes[0].DataAtom == atom.P && s.Parent().Contents().Length() == 1
		if !isParentBody && !isOnlyChild {
			log.Println(`不是 block 级别的 audio，不处理。`)
			return true
		}

		content, err := t.render(md, src)
		if err != nil {
			log.Println("渲染出错：", err)
			return true
		}

		if isParentBody {
			s.ReplaceWithHtml(content)
		} else if isOnlyChild {
			s.Parent().ReplaceWithHtml(content)
		}

		return true
	})
	return outErr
}

func (t *MediaTags) getSrc(s *goquery.Selection) string {
	src := s.AttrOr(`src`, "")
	if src == "" {
		s.Find(`source`).EachWithBreak(func(i int, s *goquery.Selection) bool {
			if value, ok := s.Attr(`src`); ok {
				src = value
				return false
			}
			return true
		})
	}
	return src
}

func (t *MediaTags) parse(src string) (tag.Metadata, error) {
	fp, err := t.web.OpenURL(src)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	fps, ok := fp.(io.ReadSeeker)
	if !ok {
		return nil, fmt.Errorf(`open 不支持定位。`)
	}

	return tag.ReadFrom(fps)
}

type Metadata struct {
	tag.Metadata
	Source   string
	FileName string
}

func (d *Metadata) PictureAsImage() template.HTML {
	if p := d.Picture(); p != nil {
		base64 := base64.RawStdEncoding.EncodeToString(p.Data)
		return template.HTML(fmt.Sprintf(`<img title="%s" src="data:%s;base64,%s">`,
			html.EscapeString(p.Description),
			p.MIMEType,
			base64,
		))
	}
	return ""
}

func (t *MediaTags) render(md tag.Metadata, source string) (string, error) {
	buf := bytes.NewBuffer(nil)
	var name string
	if u, err := url.Parse(source); err == nil {
		name = filepath.Base(u.Path)
	}
	if err := t.tmpl.GetNamed(`player.html`).Execute(buf, &Metadata{md, source, name}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
