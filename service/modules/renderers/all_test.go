package renderers_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
)

func TestMarkdown(t *testing.T) {
	tr := renderers.NewMarkdown()
	t2, s, err := tr.Render(`### <a id="my-header"></a>Header`)
	fmt.Println(t2, s, err)
}

func TestImage(t *testing.T) {
	tr := renderers.NewMarkdown()
	t2, s, err := tr.Render(`
# heading

![a](a.png)
`)
	fmt.Println(t2, s, err)
}

func TestMarkdownAll(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("test_data")))
	defer server.Close()
	host := server.URL

	cases := []struct {
		ID          float32
		Options     []renderers.Option
		Description string
		Markdown    string
		Html        string
	}{
		{
			ID:       1,
			Markdown: `![avatar](test_data/avatar.jpg?scale=0.3)`,
			Html:     `<p><img src="test_data/avatar.jpg" alt="avatar" loading="lazy" width=138 height=138 /></p>`,
		},
		{
			ID:       2,
			Markdown: `- ![avatar](test_data/avatar.jpg?scale=0.3)`,
			Html: `<ul>
<li><img src="test_data/avatar.jpg" alt="avatar" loading="lazy" width=138 height=138 /></li>
</ul>`,
		},
		{
			ID:          3,
			Description: `支持网络图片的缩放`,
			Markdown:    fmt.Sprintf(`![](%s/avatar.jpg?scale=0.1)`, host),
			Html:        fmt.Sprintf(`<p><img src="%s/avatar.jpg" alt="" loading="lazy" width=46 height=46 /></p>`, host),
		},
		{
			ID:          4,
			Description: `修改页面锚点的指向`,
			Options:     []renderers.Option{},
			Markdown:    `[A](#section)`,
			Html:        `<p><a href="#section">A</a></p>`,
		},
		{
			ID:          4.1,
			Description: `修改页面锚点的指向`,
			Options: []renderers.Option{
				renderers.WithModifiedAnchorReference("/about"),
			},
			Markdown: `[A](#section)`,
			Html:     `<p><a href="/about#section">A</a></p>`,
		},
		{
			ID:          4.2,
			Description: `修改页面锚点的指向`,
			Options: []renderers.Option{
				renderers.WithModifiedAnchorReference("/about/"),
			},
			Markdown: `[A](#section)`,
			Html:     `<p><a href="/about/#section">A</a></p>`,
		},
		{
			ID:          5.0,
			Description: `新窗口打开链接`,
			Options:     []renderers.Option{},
			Markdown:    `[](/foo)`,
			Html:        `<p><a href="/foo"></a></p>`,
		},
		{
			ID: 5.1,
			Options: []renderers.Option{
				renderers.WithOpenLinksInNewTab(),
			},
			Markdown: `[](/foo)`,
			Html:     `<p><a href="/foo" target="_blank" class="external"></a></p>`,
		},
		{
			ID: 5.1,
			Options: []renderers.Option{
				renderers.WithOpenLinksInNewTab(),
			},
			Markdown: `[](#section)`,
			Html:     `<p><a href="#section"></a></p>`,
		},
	}
	for _, tc := range cases {
		md := renderers.NewMarkdown(tc.Options...)
		if tc.ID == 2 {
			log.Println(`debug`)
		}
		_, html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(html) != strings.TrimSpace(tc.Html) {
			t.Fatalf("not equal:\n%s\n%s\n%s\n\n", tc.Markdown, tc.Html, html)
		}
	}
}
