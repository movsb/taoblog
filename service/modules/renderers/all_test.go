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
	t2, s, err := tr.Render(`$$
\sqrt{公式}
$$
`)
	fmt.Println(t2, s, err)
}

func TestMarkdownAll(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("test_data")))
	defer server.Close()
	host := server.URL

	cases := []struct {
		ID          float32
		Options     []renderers.Option2
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
			ID:       2.1,
			Markdown: `- ![avatar](test_data/中文.jpg?scale=0.3)`,
			Html: `<ul>
<li><img src="test_data/%E4%B8%AD%E6%96%87.jpg" alt="avatar" loading="lazy" width=138 height=138 /></li>
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
			Options:     []renderers.Option2{},
			Markdown:    `[A](#section)`,
			Html:        `<p><a href="#section">A</a></p>`,
		},
		{
			ID:          4.1,
			Description: `修改页面锚点的指向`,
			Options: []renderers.Option2{
				renderers.WithModifiedAnchorReference("/about"),
			},
			Markdown: `[A](#section)`,
			Html:     `<p><a href="/about#section">A</a></p>`,
		},
		{
			ID:          4.2,
			Description: `修改页面锚点的指向`,
			Options: []renderers.Option2{
				renderers.WithModifiedAnchorReference("/about/"),
			},
			Markdown: `[A](#section)`,
			Html:     `<p><a href="/about/#section">A</a></p>`,
		},
		{
			ID:          5.0,
			Description: `新窗口打开链接`,
			Options:     []renderers.Option2{},
			Markdown:    `[](/foo)`,
			Html:        `<p><a href="/foo"></a></p>`,
		},
		{
			ID: 5.1,
			Options: []renderers.Option2{
				renderers.WithOpenLinksInNewTab(renderers.OpenLinksInNewTabKindAll),
			},
			Markdown: `[](/foo)`,
			Html:     `<p><a href="/foo" class="external" target="_blank"></a></p>`,
		},
		{
			ID: 5.2,
			Options: []renderers.Option2{
				renderers.WithOpenLinksInNewTab(renderers.OpenLinksInNewTabKindAll),
			},
			Markdown: `[](#section)`,
			Html:     `<p><a href="#section"></a></p>`,
		},
		{
			ID: 6.0,
			Markdown: `
![样式1](1.jpg "样式1")
![样式2](2.jpg "样式2")
![样式3](3.jpg '样式3"><a>"')`,
			Html: `<p><img src="1.jpg" alt="样式1" title="样式1" loading="lazy" />
<img src="2.jpg" alt="样式2" title="样式2" loading="lazy" />
<img src="3.jpg" alt="样式3" title="样式3&quot;&gt;&lt;a&gt;&quot;" loading="lazy" /></p>`,
		},
		{
			ID:       7.0,
			Markdown: `![](1.png?scale=.3)`,
			Options:  []renderers.Option2{renderers.WithUseAbsolutePaths(`/911/`)},
			Html:     `<p><img src="/911/1.png" alt="" loading="lazy" /></p>`,
		},
		{
			ID:       8.0,
			Markdown: `- item`,
			Options:  []renderers.Option2{renderers.WithReserveListItemMarkerStyle()},
			Html: `<ul class="marker-minus">
<li>item</li>
</ul>`,
		},
	}
	for _, tc := range cases {
		if tc.ID == 2.0 {
			log.Println(`debug`)
		}
		options := append([]renderers.Option2{renderers.Testing()}, tc.Options...)
		md := renderers.NewMarkdown(options...)
		_, html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Fatal(err)
		}
		sep := strings.Repeat("-", 128)
		if strings.TrimSpace(html) != strings.TrimSpace(tc.Html) {
			t.Fatalf("not equal %v:\n%s\n%s\n%s\n%s\n%s\n\n", tc.ID, tc.Markdown, sep, tc.Html, sep, html)
		}
	}
}
