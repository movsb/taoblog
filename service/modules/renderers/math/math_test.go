package katex_test

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers"
	katex "github.com/movsb/taoblog/service/modules/renderers/math"
)

func TestRender(t *testing.T) {
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

func TestWASI(t *testing.T) {
	r := utils.Must1(katex.NewWebAssemblyRuntime(context.Background()))
	defer r.Close()

	html := utils.Must1(r.RenderKatex(context.Background(), `a`, false))
	t.Log(html)

	// 测试两遍以判断可能出现的模块重名、重复注册问题。
	html = utils.Must1(r.RenderKatex(context.Background(), `a`, false))
	t.Log(html)
}
