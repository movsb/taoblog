package katex

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/theme/modules/sass"
	"github.com/yuin/goldmark"
)

//go:generate sass --no-source-map static/style.scss static/style.css

//go:embed static binary katex/katex.min.css katex/style.css
var Root embed.FS

func init() {
	dynamic.RegisterInit(func() {
		katexDir := utils.Must1(fs.Sub(Root, `katex`))
		raw := utils.Must1(fs.ReadFile(katexDir, `katex/katex.min.css`))
		// 删除不必要的字体。
		stripped := regexp.MustCompile(`,url[^}]+`).ReplaceAllLiteral(raw, nil)
		customize := utils.Must1(fs.ReadFile(katexDir, `katex/style.css`))
		dynamic.Dynamic[`math`] = dynamic.Content{
			Styles: []string{
				string(stripped),
				string(customize),
			},
			Root: Root,
		}
		sass.WatchDefaultAsync(dir.SourceAbsoluteDir().Join(`katex`))
	})
}

type Math struct{}

var _ interface {
	goldmark.Extender
} = (*Math)(nil)

func New() *Math {
	return &Math{}
}

func (m *Math) Extend(md goldmark.Markdown) {
	onceRt.Do(func() {
		// TODO 关闭。
		rt = utils.Must1(NewWebAssemblyRuntime(context.TODO()))
	})

	mathjax.NewMathJax(
		mathjax.WithInlineDelim(`$`, `$`),
		mathjax.WithBlockDelim(`$$`, `$$`),
	).Extend(md)
}

func (m *Math) TransformHtml(doc *goquery.Document) error {
	process := func(s *goquery.Selection, text string, displayMode bool) {
		tex := strings.Trim(text, ` $`)
		html, err := rt.RenderKatex(context.TODO(), tex, displayMode)
		if err != nil {
			log.Println(err)
			return
		}
		s.ReplaceWithHtml(html)
	}
	doc.Find(`span.math.inline`).Each(func(_ int, s *goquery.Selection) {
		process(s, s.Text(), false)
	})
	doc.Find(`span.math.display`).Each(func(_ int, s *goquery.Selection) {
		process(s, s.Text(), true)
	})
	return nil
}

var (
	rt     *WebAssemblyRuntime
	onceRt sync.Once
)
