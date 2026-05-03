package highlight_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/highlight"
)

func TestRenderer(t *testing.T) {
	for _, tc := range []struct {
		markdown string
		html     string
	}{
		{
			markdown: "```\nabc\n```",
			html:     "<pre><code>abc\n</code></pre>",
		},
		{
			markdown: "```\n" + strings.Repeat("abc\n", 100) + "```",
			html:     `<pre><code class="gutter-3">` + strings.Repeat("abc\n", 100) + `</code></pre>`,
		},
		// {
		// 	markdown: "```sh\nabc\n```",
		// 	html:     "",
		// },
		// {
		// 	markdown: "```sh\n" + strings.Repeat("abc\n", 100) + "```",
		// 	html:     "",
		// },
	} {
		md := renderers.NewMarkdown(highlight.New())
		html, err := md.Render(tc.markdown)
		if err != nil {
			t.Error(err)
			continue
		}
		if html != tc.html {
			t.Errorf("结果不相等：\n%s\n--------\n%s", tc.html, html)
		}
	}
}
