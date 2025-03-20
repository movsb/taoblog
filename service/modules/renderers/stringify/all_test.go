package stringify_test

import (
	"log"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/emojis"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/math"
	"github.com/movsb/taoblog/service/modules/renderers/media_tags"
	"github.com/movsb/taoblog/service/modules/renderers/stringify"
	"github.com/yuin/goldmark/extension"
)

func TestPrettifier(t *testing.T) {
	cases := []struct {
		ID          float32
		Options     []any
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
			Options:  []any{},
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
		{
			ID:       9.0,
			Options:  []any{math.New()},
			Markdown: `$a$`,
			Text:     `[公式]`,
		},
		{
			ID:       10.0,
			Markdown: `[狗头]`,
			Text:     `[狗头]`, // 其实直接用表情图片本身也是很好的。
			Options: []any{
				emojis.New(utils.Must1(url.Parse(`/`))),
			},
		},
		{
			ID: 11.0,
			Options: []any{
				media_tags.New(gold_utils.NewWebFileSystem(os.DirFS(`../media_tags/test_data`), &url.URL{Path: `/`})),
			},
			Markdown: `
万物死

<audio controls>
	<source src="杨晚晚 - 片片相思赋予谁.mp3" />
</audio>
`,
			Text: `万物死
[音乐]`,
		},
	}
	for _, tc := range cases {
		options := append(tc.Options, extension.GFM, extension.Footnote)
		options = append(options, renderers.WithHtmlPrettifier(stringify.New()))
		md := renderers.NewMarkdown(options...)
		if tc.ID == 9 {
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
