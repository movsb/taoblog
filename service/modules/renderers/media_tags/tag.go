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
)

//go:embed player.html
var _root embed.FS

type MediaTags struct {
	root fs.FS
	tmpl *utils.TemplateLoader
}

var onceTmpl sync.Once
var _tmpl *utils.TemplateLoader

func New(root fs.FS, tmpl *utils.TemplateLoader) *MediaTags {
	if tmpl == nil {
		onceTmpl.Do(func() {
			_tmpl = utils.NewTemplateLoader(_root, nil, nil)
		})
		tmpl = _tmpl
	}
	return &MediaTags{
		root: root,
		tmpl: tmpl,
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
		u, err := url.Parse(src)
		if err != nil {
			log.Println(`路径解析错误：`, err)
			return true
		}
		if u.IsAbs() {
			log.Println(`不是相对路径不解析。`, u.String(), u.Path)
			return true
		}
		md, err := t.parse(u.Path)
		if err != nil {
			// 忽略错误，继续
			log.Println("解析数据出错：", err)
			return true
		}
		if err := t.render(s, md, u.Path); err != nil {
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
	fp, err := t.root.Open(src)
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
	name := filepath.Base(source)
	if err := t.tmpl.GetNamed(`player.html`).Execute(buf, &Metadata{md, source, utils.RandomString(), name}); err != nil {
		return err
	}
	s.ReplaceWithHtml(buf.String())
	return nil
}
