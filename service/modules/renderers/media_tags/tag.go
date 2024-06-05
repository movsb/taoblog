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
	"log"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dhowden/tag"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
)

var SourceRelativeDir = dir.SourceRelativeDir()

//go:embed player.html
var _root embed.FS

type MediaTags struct {
	web  gold_utils.WebFileSystem
	tmpl *utils.TemplateLoader

	devMode      bool
	themeChanged func()
}

var _gOnceTmpl sync.Once
var _gTmpl *utils.TemplateLoader

type Option func(*MediaTags)

func WithDevMode(themeChanged func()) Option {
	return func(mt *MediaTags) {
		mt.devMode = true
		mt.themeChanged = themeChanged
	}
}

func New(web gold_utils.WebFileSystem, options ...Option) *MediaTags {
	tag := &MediaTags{
		web: web,
	}

	for _, opt := range options {
		opt(tag)
	}

	// 判断为空的目的是测试里面可能会预初始化。
	if _gTmpl == nil {
		if tag.devMode {
			_gOnceTmpl.Do(func() {
				_gTmpl = utils.NewTemplateLoader(_root, nil, nil)
			})
		} else {
			_gTmpl = utils.NewTemplateLoader(utils.NewDirFSWithNotify(string(SourceRelativeDir)), nil, tag.themeChanged)
		}
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
		//
		// 所以目前处理所以类型的 audio，而不仅是 block 级别。
		// 仅当是唯一子元素（即 block）时，才渲染。
		// if s.Parent().Contents().Length() != 1 {
		// 	log.Println(`不是 block 级别的 audio，不处理。`)
		// 	return true
		// }

		if err := t.render(s, md, src); err != nil {
			log.Println("渲染出错：", err)
			return true
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
	Random   string
	FileName string
}

func (d *Metadata) PictureAsImage() template.HTML {
	if p := d.Picture(); p != nil {
		base64 := base64.RawStdEncoding.EncodeToString(p.Data)
		return template.HTML(fmt.Sprintf(`<img title="%s" src="data:%s;base64,%s" />`,
			html.EscapeString(p.Description),
			p.MIMEType,
			base64,
		))
	}
	return ""
}

func (t *MediaTags) render(s *goquery.Selection, md tag.Metadata, source string) error {
	buf := bytes.NewBuffer(nil)
	var name string
	if u, err := url.Parse(source); err == nil {
		name = filepath.Base(u.Path)
	}
	if err := t.tmpl.GetNamed(`player.html`).Execute(buf, &Metadata{md, source, utils.RandomString(), name}); err != nil {
		return err
	}
	rendered := buf.String()
	// 由于渲染结果是 div，并且 audio 属于 inline 元素，所以需要去掉父元素。
	s.Parent().ReplaceWithHtml(rendered)
	return nil
}
