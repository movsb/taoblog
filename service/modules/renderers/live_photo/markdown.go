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
	"golang.org/x/net/html/atom"
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
	return utils.NewTemplateLoader(utils.IIF(version.DevMode(), _local, fs.FS(_embed)), nil, dynamic.Reload)
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

		// [!NOTE]
		//
		// 1. 由于渲染出来有 div 元素，需要把上一级的 p 替换掉。
		// 2. 由于 Live Photo 一般应保持原始尺寸，所以只处理单张图片的时候。
		if p := s.Nodes[0].Parent; p != nil && p.DataAtom == atom.P {
			if self := s.Nodes[0]; self.PrevSibling == nil && self.NextSibling == nil {
				html := e.render(s, width, height, video)
				s.Parent().ReplaceWithHtml(string(html))
			}
		}
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

	checkExt := func(ext string) (string, bool) {
		videoPath := pathNoExt + ext
		fp, err := e.web.OpenURL(videoPath)
		if err != nil {
			return ``, false
		}
		fp.Close()
		return videoPath, true
	}

	for _, ext := range []string{`.mp4`, `.webm`} {
		if u, ok := checkExt(ext); ok {
			return u
		}
	}

	return ``
}

type Data struct {
	Width, Height int
	ImageElement  template.HTML
	Video         string
	Ratio         float32
}

func (e *LivePhoto) render(s *goquery.Selection, width, height int, video string) []byte {
	buf := bytes.NewBuffer(nil)
	goquery.Render(buf, s)
	d := Data{
		Width:        width,
		Height:       height,
		ImageElement: template.HTML(buf.Bytes()),
		Video:        video,
		Ratio:        float32(width) / float32(height),
	}
	buf.Reset()
	t().GetNamed(`template.html`).Execute(buf, d)
	return buf.Bytes()
}
