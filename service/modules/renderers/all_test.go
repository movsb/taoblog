package renderers_test

import (
	"compress/gzip"
	"context"
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
	"github.com/movsb/taoblog/service/modules/renderers/encrypted"
	"github.com/movsb/taoblog/service/modules/renderers/gallery"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/hashtags"
	"github.com/movsb/taoblog/service/modules/renderers/image"
	"github.com/movsb/taoblog/service/modules/renderers/lazy"
	"github.com/movsb/taoblog/service/modules/renderers/link_target"
	"github.com/movsb/taoblog/service/modules/renderers/list_markers"
	"github.com/movsb/taoblog/service/modules/renderers/live_photo"
	"github.com/movsb/taoblog/service/modules/renderers/media_size"
	"github.com/movsb/taoblog/service/modules/renderers/page_link"
	"github.com/movsb/taoblog/service/modules/renderers/scoped_css"
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
	server := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer server.Close()
	host := server.URL

	// for hashtags test
	var outputTags []string

	testDataURL, _ := url.Parse(`/`)
	web := gold_utils.NewWebFileSystem(os.DirFS(`testdata`), testDataURL)

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
			Options:  []renderers.Option2{media_size.New(web)},
			Html:     `<p><img src="avatar.jpg" alt="avatar" width="138" height="138"/></p>`,
		},
		{
			ID:       2,
			Markdown: `- ![avatar](avatar.jpg?scale=0.3)`,
			Options:  []renderers.Option2{media_size.New(web)},
			Html: `<ul>
<li><img src="avatar.jpg" alt="avatar" width="138" height="138"/></li>
</ul>`,
		},
		{
			ID:       2.1,
			Markdown: `- ![avatar](中文.jpg?scale=0.3)`,
			Options:  []renderers.Option2{media_size.New(web)},
			Html: `<ul>
<li><img src="%E4%B8%AD%E6%96%87.jpg" alt="avatar" width="138" height="138"/></li>
</ul>`,
		},
		{
			ID:          3,
			Description: `支持网络图片的缩放`,
			Markdown:    fmt.Sprintf(`![](%s/avatar.jpg?scale=0.1)`, host),
			Options:     []renderers.Option2{media_size.New(web)},
			Html:        fmt.Sprintf(`<p><img src="%s/avatar.jpg" alt="" width="46" height="46"/></p>`, host),
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
				link_target.New(link_target.OpenLinksInNewTabKindAll),
			},
			Markdown: `[](/foo)`,
			Html:     `<p><a href="/foo" class="external" target="_blank"></a></p>`,
		},
		{
			ID: 5.2,
			Options: []renderers.Option2{
				link_target.New(link_target.OpenLinksInNewTabKindAll),
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
			Html: `<p><img src="1.jpg" alt="样式1" title="样式1">
<img src="2.jpg" alt="样式2" title="样式2">
<img src="3.jpg" alt="样式3" title="样式3&quot;&gt;&lt;a&gt;&quot;"></p>`,
		},
		{
			ID:       8.0,
			Markdown: `- item`,
			Options:  []renderers.Option2{list_markers.New()},
			Html: `<ul class="marker-minus">
<li>item</li>
</ul>`,
		},
		{
			// https://blog.twofei.com/1706/#comment-2077
			ID: 9.0,
			Markdown: `
<iframe></iframe>

<img>

<audio></audio>

<video></video>

<video preload="metadata"></video>
`,
			Options: []renderers.Option2{
				image.New(nil),
				lazy.New()},
			// 奇怪🤔，为什么 <img> 不会被放在 <p> 中？
			Html: `
<iframe loading="lazy"></iframe>
<img loading="lazy"/>
<p><audio preload="none"></audio></p>
<p><video preload="none"></video></p>
<p><video preload="metadata"></video></p>
`,
		},
		{
			ID:      10.0,
			Options: []renderers.Option2{gallery.New()},
			Markdown: `<Gallery>

![](1.jpg)
![](2.jpg)

</Gallery>`,
			Html: `<div class="gallery"><img src="1.jpg" alt=""/><img src="2.jpg" alt=""/></div>`,
		},
		{
			ID: 11.0,
			Options: []renderers.Option2{hashtags.New(func(tag string) string {
				return utils.DropLast1(url.Parse(`http://localhost/tags`)).JoinPath(tag).String()
			}, &outputTags)},
			Markdown: `a #b c #桃子`,
			Html:     `<p>a <span class="hashtag"><a href="http://localhost/tags/b">#b</a></span> c <span class="hashtag"><a href="http://localhost/tags/%E6%A1%83%E5%AD%90">#桃子</a></span></p>`,
		},
		{
			ID:       12.0,
			Options:  []renderers.Option2{scoped_css.New(`article#123`)},
			Markdown: `<style>table { min-width: 100px; }</style>`,
			Html:     `<style>article#123 table{min-width:100px;}</style>`,
		},
		{
			ID: 13.0,
			Options: []renderers.Option2{
				media_size.New(web),
				live_photo.New(context.Background(), web),
				encrypted.New(),
			},
			Markdown: `![](1.jpg)`,
			Html: `
<div><div class="live-photo" style="width: 460px; height: 460px; aspect-ratio: 1;">
	<div class="container">
		<video src="1.mp4" playsinline="" onerror="decryptFile(this)" onended="fixVideoCache(this)"></video>
		<img src="1.jpg" alt="" width="460" height="460" onerror="decryptFile(this)"/>
	</div>
	<div class="icon">
		<img src="/v3/dynamic/live-photo/live.png" class="static"/>
		<span>实况</span>
	</div>
	<div class="warning" style="opacity: 0;"></div>
</div>
</div>`,
		},
		{
			ID:          14.0,
			Description: `表情和页面嵌入同时使用的时候渲染出错`,
			Options: []renderers.Option2{
				emojis.New(emojis.BaseURLForDynamic),
				page_link.New(context.Background(), func(_ context.Context, id int32) (string, error) {
					return fmt.Sprint(id), nil
				}, nil),
			},
			Markdown: `[狗头] A [[123]]`,
			Html:     `<p><img src="/v3/dynamic/emojis/weixin/doge.png" alt="[狗头]" title="狗头" class="static emoji weixin"/> A <a href="/123/">123</a></p>`,
		},
		{
			ID:          15.0,
			Description: `连续的表情被当成Gallery`,
			Options: []renderers.Option2{
				gallery.New(),
				emojis.New(emojis.BaseURLForDynamic),
			},
			Markdown: `[doge][doge]`,
			Html:     `<p><img src="/v3/dynamic/emojis/weixin/doge.png" alt="[doge]" title="doge" class="static emoji weixin"/><img src="/v3/dynamic/emojis/weixin/doge.png" alt="[doge]" title="doge" class="static emoji weixin"/></p>`,
		},
	}
	for _, tc := range cases {
		if tc.ID == 14.0 {
			log.Println(`debug`)
		}
		options := []renderers.Option2{}
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

func TestSpec(t *testing.T) {
	type _Case struct {
		Example  int    `yaml:"example"`
		Markdown string `yaml:"markdown"`
		HTML     string `yaml:"html"`

		StartLine int    `yaml:"start_line"`
		EndLine   int    `yaml:"end_line"`
		Section   string `yaml:"section"`
	}

	fp := utils.Must1(os.Open(`testdata/spec-0.31.2.json.gz`))
	defer fp.Close()
	gr := utils.Must1(gzip.NewReader(fp))

	testCases := test_utils.MustLoadCasesFromYamlReader[_Case](gr)
	baseURL, _ := url.Parse(`/`)
	for _, tc := range testCases {
		md := renderers.NewMarkdown(
			emojis.New(baseURL),
			page_link.New(context.Background(), func(ctx context.Context, id int32) (string, error) {
				return fmt.Sprint(id), nil
			}, nil),
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

func TestTitle(t *testing.T) {
	testCases := []struct {
		Markdown string
		Title    string
	}{
		{
			Markdown: `# 123`,
			Title:    `123`,
		},
		{
			Markdown: `# \123`,
			Title:    `\123`,
		},
		{
			Markdown: "# `123`",
			Title:    `123`,
		},
		{
			Markdown: `# \<img\>`,
			Title:    `<img>`,
		},
		{
			Markdown: `# <img>1`,
			Title:    `<img>1`,
		},
		{
			Markdown: `# &lt;X`,
			Title:    `<X`,
		},
	}

	for i, tc := range testCases {
		var title string
		md := renderers.NewMarkdown(renderers.WithTitle(&title))
		md.Render(tc.Markdown)
		if title != tc.Title {
			t.Errorf(`标题不相等：#%d: [%s] vs. [%s]`, i, tc.Title, title)
			continue
		}
	}
}
