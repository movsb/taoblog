package katex

import (
	"bytes"
	"embed"
	"encoding/json"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/movsb/taoblog/modules/utils"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/yuin/goldmark"
)

//go:embed fonts katex.min.css style.css
var _root embed.FS

func init() {
	raw := utils.Must1(_root.ReadFile(`katex.min.css`))
	// 删除不必要的字体。
	stripped := regexp.MustCompile(`,url[^}]+`).ReplaceAllLiteral(raw, nil)
	customize := utils.Must1(_root.ReadFile(`style.css`))
	dynamic.Dynamic[`math`] = dynamic.Content{
		Styles: []string{
			string(stripped),
			string(customize),
		},
		Root: _root,
	}
}

type Math struct{}

var _ interface {
	goldmark.Extender
} = (*Math)(nil)

func New() *Math {
	return &Math{}
}

func (m *Math) Extend(md goldmark.Markdown) {
	mathjax.NewMathJax(
		mathjax.WithInlineDelim(`$`, `$`),
		mathjax.WithBlockDelim(`$$`, `$$`),
	).Extend(md)
}

func (m *Math) TransformHtml(doc *goquery.Document) error {
	process := func(s *goquery.Selection, text string, displayMode bool) {
		tex := strings.Trim(text, ` $`)
		html, err := render(tex, displayMode)
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

type Options struct {
	Tex         string `json:"tex"`
	DisplayMode bool   `json:"displayMode"`
}

// TODO 缓存结果。
func render(tex string, displayMode bool) (string, error) {
	args := Options{
		Tex:         tex,
		DisplayMode: displayMode,
	}
	body, _ := json.Marshal(args)
	cmd := exec.Command(`./katex`)
	cmd.Stdin = bytes.NewReader(body)
	out, err := cmd.Output()
	return string(out), err
}
