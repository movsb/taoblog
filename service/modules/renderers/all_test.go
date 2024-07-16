package renderers_test

import (
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	test_utils "github.com/movsb/taoblog/modules/utils/test"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/emojis"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
	"github.com/movsb/taoblog/service/modules/renderers/imaging"
	"github.com/movsb/taoblog/service/modules/renderers/media_size"
	"github.com/yuin/goldmark/extension"
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

	// for hashtags test
	var outputTags []string

	testDataURL, _ := url.Parse(`/`)

	cases := []struct {
		ID          float32
		Options     []renderers.Option2
		Description string
		Markdown    string
		Html        string
	}{
		{
			ID:       1,
			Markdown: `![avatar](avatar.jpg?scale=0.3)`,
			Options:  []renderers.Option2{media_size.New(gold_utils.NewWebFileSystem(os.DirFS("test_data"), testDataURL))},
			Html:     `<p><img src="avatar.jpg" alt="avatar" loading="lazy" width="138" height="138"/></p>`,
		},
		{
			ID:       2,
			Markdown: `- ![avatar](avatar.jpg?scale=0.3)`,
			Options:  []renderers.Option2{media_size.New(gold_utils.NewWebFileSystem(os.DirFS("test_data"), testDataURL))},
			Html: `<ul>
<li><img src="avatar.jpg" alt="avatar" loading="lazy" width="138" height="138"/></li>
</ul>`,
		},
		{
			ID:       2.1,
			Markdown: `- ![avatar](中文.jpg?scale=0.3)`,
			Options:  []renderers.Option2{media_size.New(gold_utils.NewWebFileSystem(os.DirFS("test_data"), testDataURL))},
			Html: `<ul>
<li><img src="%E4%B8%AD%E6%96%87.jpg" alt="avatar" loading="lazy" width="138" height="138"/></li>
</ul>`,
		},
		{
			ID:          3,
			Description: `支持网络图片的缩放`,
			Markdown:    fmt.Sprintf(`![](%s/avatar.jpg?scale=0.1)`, host),
			Options:     []renderers.Option2{media_size.New(gold_utils.NewWebFileSystem(os.DirFS("test_data"), testDataURL))},
			Html:        fmt.Sprintf(`<p><img src="%s/avatar.jpg" alt="" loading="lazy" width="46" height="46"/></p>`, host),
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
			Html: `<p><img src="1.jpg" alt="样式1" title="样式1" loading="lazy"/>
<img src="2.jpg" alt="样式2" title="样式2" loading="lazy"/>
<img src="3.jpg" alt="样式3" title="样式3&quot;&gt;&lt;a&gt;&quot;" loading="lazy"/></p>`,
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
		{
			ID:      10.0,
			Options: []renderers.Option2{imaging.WithGallery()},
			Markdown: `<Gallery>

![](1.jpg)
![](2.jpg)

</Gallery>`,
			Html: `<div class="gallery"><img src="1.jpg" alt="" loading="lazy"/><img src="2.jpg" alt="" loading="lazy"/></div>`,
		},
		{
			ID: 11.0,
			Options: []renderers.Option2{renderers.WithHashTags(func(tag string) string {
				return utils.DropLast1(url.Parse(`http://localhost/tags`)).JoinPath(tag).String()
			}, &outputTags)},
			Markdown: `a #b c #桃子`,
			Html:     `<p>a <span class="hashtag"><a href="http://localhost/tags/b">#b</a></span> c <span class="hashtag"><a href="http://localhost/tags/%E6%A1%83%E5%AD%90">#桃子</a></span></p>`,
		},
	}
	for _, tc := range cases {
		if tc.ID == 10.0 {
			log.Println(`debug`)
		}
		options := []renderers.Option2{renderers.WithImageRenderer()}
		options = append(options, tc.Options...)
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
		{
			ID:       6,
			Markdown: "一直用 `@media screen`，今天才知道有 [`@container`][mdn] 这么个神器\n\n[mdn]: http://mdn",
			Text:     `一直用 @media screen，今天才知道有 @container 这么个神器`,
		},
		{
			ID:       7.1,
			Markdown: `[https://blog.twofei.com/118/%e4%b8%87%e7%89%a9%e6%ad%bb%208-bit.mp3](https://blog.twofei.com/118/%e4%b8%87%e7%89%a9%e6%ad%bb%208-bit.mp3)`,
			Text:     `[链接]`,
		},
		{
			ID:       7.2,
			Markdown: `[万物死](https://blog.twofei.com/118/%e4%b8%87%e7%89%a9%e6%ad%bb%208-bit.mp3)`,
			Text:     `万物死`,
		},
		{
			ID:       8.0,
			Markdown: "用 `<script>` 嵌入 JSON 的正规做法[^1]：\n\n[^1]: https://",
			Text:     `用 <script> 嵌入 JSON 的正规做法：`,
		},
		// TODO: 测试的时候 katex 还没 build 出来，暂时跑不了。
		// {
		// 	ID:       9.0,
		// 	Options:  []renderers.Option2{katex.New()},
		// 	Markdown: `$a$`,
		// 	Text:     `[公式]`,
		// },
	}
	for _, tc := range cases {
		options := append(tc.Options, renderers.WithHtmlPrettifier(), extension.GFM, extension.Footnote)
		md := renderers.NewMarkdown(options...)
		if tc.ID == 9.0 {
			log.Println(`debug`)
		}
		text, err := md.Render(tc.Markdown)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(text) != strings.TrimSpace(tc.Text) {
			t.Fatalf("not equal %v:\nMarkdown:%s\nExpected:%s\nGot     :%s\n\n", tc.ID, tc.Markdown, tc.Text, text)
		}
	}
}

func TestSpec(t *testing.T) {
	type _Case struct {
		Example  int    `yaml:"example"`
		Markdown string `yaml:"markdown"`
		HTML     string `yaml:"html"`
	}

	fp := utils.Must1(os.Open(`test_data/spec-0.31.2.json.gz`))
	defer fp.Close()
	gr := utils.Must1(gzip.NewReader(fp))

	testCases := test_utils.MustLoadCasesFromYamlReader[_Case](gr)
	for _, tc := range testCases {
		md := renderers.NewMarkdown(
			emojis.New(),
			renderers.WithXHTML(),
			renderers.WithoutTransform(),
		)
		h, err := md.Render(tc.Markdown)
		if err != nil {
			t.Error(err)
			continue
		}
		if h != tc.HTML {
			t.Errorf("example %d error:\nmd  : %s\nwant: %s\ngot : %s\n", tc.Example, tc.Markdown, tc.HTML, h)
			continue
		}
	}
}
