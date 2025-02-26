package katex_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	katex "github.com/movsb/taoblog/service/modules/renderers/math"
)

func TestRender(t *testing.T) {
	t.SkipNow()
	tests := []struct {
		Markdown string
		Html     string
	}{
		{
			Markdown: `$a$`,
			Html:     `<p><span class="math inline">$a$</span></p>`,
		},
		{
			Markdown: `$$a$$`,
			Html:     `<p><span class="math display">$$a$$</span></p>`,
		},
	}

	for i, tt := range tests {
		md := renderers.NewMarkdown(katex.New(), renderers.WithoutTransform())
		html, err := md.Render(tt.Markdown)
		if err != nil {
			t.Error(i, err)
			continue
		}
		if strings.TrimSpace(html) != tt.Html {
			t.Errorf("不相等：\nwant:\n%s\n got:\n%s", tt.Html, html)
		}
	}
}
