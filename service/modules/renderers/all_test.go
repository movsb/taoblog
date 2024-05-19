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
	s, err := tr.Render(`$$
\sqrt{公式}
$$
`)
	fmt.Println(s, err)
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
			Html:     `<p><img src="/911/1.png" alt="" loading="lazy"/></p>`,
		},
		{
			ID: 7.1,
			Markdown: `<audio><source src="1.mp3"/></audio>
<video><source src="1.mp4"/></video>
<iframe src="1.html"></iframe>
<object data="1.pdf"></object>
`,
			Options: []renderers.Option2{renderers.WithUseAbsolutePaths(`/911/`)},
			Html: `<p><audio><source src="/911/1.mp3"/></audio>
<video><source src="/911/1.mp4"/></video></p>
<iframe src="/911/1.html"></iframe>
<object data="/911/1.pdf"></object>
`,
		},
		{
			ID:       8.0,
			Markdown: `- item`,
			Options:  []renderers.Option2{renderers.WithReserveListItemMarkerStyle()},
			Html: `<ul class="marker-minus">
<li>item</li>
</ul>`,
		},
		{
			ID:       9.0,
			Markdown: `<iframe width="560" height="315" src="https://www.youtube.com/embed/7FiQV1-z06Q?si=013GZ9Dja-o8n2EY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>`,
			Options:  []renderers.Option2{renderers.WithLazyLoadingFrames()},
			Html:     `<iframe width="560" height="315" src="https://www.youtube.com/embed/7FiQV1-z06Q?si=013GZ9Dja-o8n2EY" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen="" loading="lazy"></iframe>`,
		},
	}
	for _, tc := range cases {
		if tc.ID == 7.1 {
			log.Println(`debug`)
		}
		options := append([]renderers.Option2{renderers.Testing()}, tc.Options...)
		md := renderers.NewMarkdown(options...)
		html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Fatal(tc.ID, err)
		}
		sep := strings.Repeat("-", 128)
		if strings.TrimSpace(html) != strings.TrimSpace(tc.Html) {
			t.Fatalf("not equal %v:\n%s\n%s\n%s\n%s\n%s\n\n", tc.ID, tc.Markdown, sep, tc.Html, sep, html)
		}
	}
}

func TestPrettifier(t *testing.T) {
	cases := []struct {
		ID          float32
		Options     []renderers.Option2
		Description string
		Markdown    string
		Text        string
	}{
		{
			ID:       1,
			Markdown: `![文本](/URL)`,
			Text:     `[图片]`,
		},
		{
			ID:       2,
			Markdown: `[文本](链接)`,
			Text:     `文本`,
		},
		{
			ID:       2.1,
			Markdown: `[https://a](https://a)`,
			Text:     `[链接]`,
		},
		{
			ID:       2.2,
			Markdown: `<https://a>`,
			Text:     `[链接]`,
		},
		{
			ID:       3,
			Options:  []renderers.Option2{},
			Markdown: "代码：\n\n```go\npackage main\n```",
			Text:     "代码：\n[代码]",
		},
		{
			ID: 4,
			Markdown: `|h1|h2|
|-|-|
|1|2|
`,
			Text: `[表格]`,
		},
		{
			ID: 5,
			Markdown: `
光看歌词怎么行？必须种草易仁莲：

<iframe></iframe>
<video></video>
<audio></audio>
<canvas></canvas>
<embed></embed>
<map></map>
<object></object>
<script></script>
<svg></svg>
<code></code>
`,
			Text: `光看歌词怎么行？必须种草易仁莲：
[页面]
[视频]
[音频]
[画布]
[对象]
[地图]
[对象]
[脚本]
[图片]
[代码]
`,
		},
	}
	for _, tc := range cases {
		options := append([]renderers.Option2{renderers.Testing()}, tc.Options...)
		options = append(options, renderers.WithHtmlPrettifier())
		md := renderers.NewMarkdown(options...)
		if tc.ID == 6.0 {
			log.Println(`debug`)
		}
		text, err := md.Render(tc.Markdown)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(text) != strings.TrimSpace(tc.Text) {
			t.Fatalf("not equal %v:\nMarkdown:%s\nExpected:%s\nGot:%s\n\n", tc.ID, tc.Markdown, tc.Text, text)
		}
	}
}
