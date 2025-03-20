package friends

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/url"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"gopkg.in/yaml.v2"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed friend.html style.css
var _embed embed.FS
var _root = os.DirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `friends`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

type Friends struct {
	task   *Task
	postID int
}

type Option func(f *Friends)

func New(task *Task, postID int, options ...Option) *Friends {
	f := &Friends{
		task:   task,
		postID: postID,
	}

	for _, opt := range options {
		opt(f)
	}

	return f
}

type Friend struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Description string `yaml:"description"`

	// 日夜共用 & 日夜分开
	Icon  string    `yaml:"icon"`
	Icons [2]string `yaml:"icons"`

	iconDataURLs [2]string
}

func (f Friend) LightDataURL() template.URL {
	return template.URL(f.iconDataURLs[0])
}

func (f Friend) DarkDataURL() template.URL {
	return template.URL(f.iconDataURLs[1])
}

var t = utils.NewTemplateLoader(utils.IIF(version.DevMode(), _root, fs.FS(_embed)), nil, func() {})

func (f *Friends) TransformHtml(doc *goquery.Document) error {
	// TODO 这个写法太泛了，容易 match 到意外的东西。
	list := doc.Find(`script[type="application/yaml"]`)
	if list.Length() == 0 {
		return nil
	}
	content := list.Nodes[0].FirstChild.Data
	var bundle struct {
		Friends []*Friend `yaml:"friends"`
	}
	if err := yaml.Unmarshal([]byte(content), &bundle); err != nil {
		log.Println(err)
		return nil
	}

	f.prepareIconURL(bundle.Friends)

	doc.Find(`div.friends`).Each(func(_ int, s *goquery.Selection) {
		for _, f := range bundle.Friends {
			buf := bytes.NewBuffer(nil)
			utils.Must(t.GetNamed(`friend.html`).Execute(buf, f))
			s.AppendHtml(buf.String())
		}
	})

	return nil
}

var svg = `<svg width="80" height="80" xmlns="http://www.w3.org/2000/svg">
 <g class="layer">
  <text font-size="50" fill="#BB0000" text-anchor="middle" x="40" y="58">%s</text>
 </g>
</svg>
`

// Name 是网站的名字，取首字符的大写作为图标。
func genSvgURL(name string) string {
	var first rune
	if len(name) > 0 {
		first, _ = utf8.DecodeRune([]byte(strings.ToUpper(name)))
	}
	letter := fmt.Sprintf(svg, string(first))
	return `data:image/svg+xml;base64,` + base64.StdEncoding.EncodeToString([]byte(letter))
}

// TODO 更好的做法是 parse 页面，取 <link rel="favicon" ...
func resolveIconURL(siteURL, faviconURL string) (string, error) {
	us, err := url.Parse(siteURL)
	if err != nil {
		return ``, err
	}

	// 如果没指定，默认用 favicon.ico
	if faviconURL == "" {
		faviconURL = us.JoinPath(`/favicon.ico`).String()
	}

	uf, err := url.Parse(faviconURL)
	if err != nil {
		return ``, err
	}

	uf = us.ResolveReference(uf)
	return uf.String(), nil
}

func (f *Friends) prepareIconURL(fs []*Friend) {
	for _, fr := range fs {
		if fr.Icons[0] == `` {
			fr.Icons[0] = fr.Icon
		}
		for i := 0; i < 2; i++ {
			if i == 1 && fr.Icons[i] == `` {
				continue
			}
			u, err := resolveIconURL(fr.URL, fr.Icons[i])
			if err != nil {
				log.Println(err)
				continue
			}
			contentType, content, found := f.task.Get(f.postID, u)
			if !found {
				// 预填充成 SVG 首字母（因为可能加载失败）。
				fr.iconDataURLs[i] = genSvgURL(fr.Name)
			} else {
				fr.iconDataURLs[i] = fmt.Sprintf(`data:%s;base64,%s`, contentType, base64.StdEncoding.EncodeToString(content))
			}
		}
	}
}
