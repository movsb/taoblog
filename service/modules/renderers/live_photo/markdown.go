package live_photo

import (
	"bytes"
	"context"
	"embed"
	"html/template"
	"io/fs"
	"log"
	urlpkg "net/url"
	pathpkg "path"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
)

//go:embed style.css script.js template.html
var _embed embed.FS
var _local = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

//go:embed live.png
var _pubEmbed embed.FS

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `live-photo`
		dynamic.WithRoots(module, _pubEmbed, _local, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
		dynamic.WithScripts(module, `script.js`)
	})
}

type LivePhoto struct {
	ctx context.Context
	web gold_utils.WebFileSystem
}

func New(ctx context.Context, web gold_utils.WebFileSystem) *LivePhoto {
	return &LivePhoto{
		ctx: ctx,
		web: web,
	}
}

var t = sync.OnceValue(func() *utils.TemplateLoader {
	return utils.NewTemplateLoader(utils.IIF(version.DevMode(), _local, fs.FS(_embed)), nil, func() {})
})

func (e *LivePhoto) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img`).Each(func(i int, s *goquery.Selection) {
		width, _ := strconv.Atoi(s.AttrOr(`width`, `0`))
		height, _ := strconv.Atoi(s.AttrOr(`height`, `0`))
		if width <= 0 || height <= 0 {
			return
		}
		video := e.checkHasVideo(s)
		if video == `` {
			return
		}
		html := e.render(s, width, height, video)
		s.ReplaceWithHtml(string(html))
	})
	return nil
}

func (e *LivePhoto) checkHasVideo(s *goquery.Selection) (video string) {
	src := s.AttrOr(`src`, ``)
	if src == `` {
		return
	}
	url, err := urlpkg.Parse(src)
	if err != nil {
		log.Println(`LivePhoto:`, err)
		return
	}
	if url.IsAbs() || url.Host != `` {
		return
	}
	path := url.Path
	ext := pathpkg.Ext(path)
	pathNoExt, _ := strings.CutSuffix(path, ext)
	videoPath := pathNoExt + `.mp4` // 硬编码的
	fp, err := e.web.OpenURL(videoPath)
	if err != nil {
		return
	}
	fp.Close()
	return videoPath
}

type Data struct {
	Width, Height int
	ImageElement  template.HTML
	Video         string
}

func (e *LivePhoto) render(s *goquery.Selection, width, height int, video string) []byte {
	buf := bytes.NewBuffer(nil)
	goquery.Render(buf, s)
	d := Data{
		Width:        width,
		Height:       height,
		ImageElement: template.HTML(buf.Bytes()),
		Video:        video,
	}
	buf.Reset()
	t().GetNamed(`template.html`).Execute(buf, d)
	return buf.Bytes()
}
