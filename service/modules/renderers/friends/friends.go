package friends

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"gopkg.in/yaml.v2"
)

//go:embed friend.html
var _root embed.FS

type Friends struct {
}

type Option func(f *Friends)

func New(options ...Option) *Friends {
	f := &Friends{}

	for _, opt := range options {
		opt(f)
	}

	return f
}

type Friend struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Description string `yaml:"description"`
	Icon        string `yaml:"icon"`
	iconDataURL string
}

func (f Friend) IconURL() template.URL {
	if f.iconDataURL != "" {
		return template.URL(f.iconDataURL)
	}
	if f.Icon != "" {
		return template.URL(f.Icon)
	}
	return ""
}

var tmpl = template.Must(template.New(`friend`).Parse(string(utils.Must1(_root.ReadFile(`friend.html`)))))

func (f *Friends) TransformHtml(doc *goquery.Document) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	f.prepareIconURL(ctx, bundle.Friends)
	<-ctx.Done()

	doc.Find(`div.friends`).Each(func(_ int, s *goquery.Selection) {
		for _, f := range bundle.Friends {
			buf := bytes.NewBuffer(nil)
			utils.Must(tmpl.Execute(buf, f))
			s.AppendHtml(buf.String())
		}
	})

	return nil
}

var svg = `<svg width="80" height="80" xmlns="http://www.w3.org/2000/svg">
 <g class="layer">
  <text font-size="50" text-anchor="middle" x="40" y="58">%s</text>
 </g>
</svg>
`

func (f *Friends) prepareIconURL(ctx context.Context, fs []*Friend) {
	for i, fr := range fs {
		if fr.Icon == "" {
			u2, err := url.Parse(fr.URL)
			if err == nil {
				if u2.Path == "" {
					u2.Path = "/"
				}
				fr.Icon = u2.JoinPath(`/favicon.ico`).String()
			}
		}
		if fr.Icon == "" {
			continue
		}
		go func(i int, fr *Friend) {
			defer func() {
				if fr.iconDataURL == "" {
					var first rune
					if len(fr.Name) > 0 {
						first, _ = utf8.DecodeRune([]byte(fr.Name))
					}
					letter := fmt.Sprintf(svg, string(first))
					fr.iconDataURL = `data:image/svg+xml;base64,` + base64.StdEncoding.EncodeToString([]byte(letter))
				}
			}()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, fr.Icon, nil)
			if err != nil {
				return
			}
			rsp, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			defer rsp.Body.Close()
			if rsp.StatusCode != http.StatusOK {
				return
			}
			body, _ := io.ReadAll(io.LimitReader(rsp.Body, 100<<10))
			contentType, _, _ := mime.ParseMediaType(rsp.Header.Get(`Content-Type`))
			if contentType == "" {
				contentType = http.DetectContentType(body)
			}
			if contentType == "" {
				return
			}
			uri := fmt.Sprintf(`data:%s;base64,%s`, contentType, base64.StdEncoding.EncodeToString(body))
			fr.iconDataURL = uri
		}(i, fr)
	}
}
