package page_link_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/page_link"
)

func TestRender(t *testing.T) {
	titles := map[int32]string{
		123: "123",
		111: "<p>",
	}
	getPageTitle := func(ctx context.Context, id int32) (string, error) {
		if t, ok := titles[id]; ok {
			return t, nil
		}
		return ``, fmt.Errorf(`id 不存在：%d`, id)
	}

	tests := []struct {
		markdown string
		html     string
	}{
		{
			markdown: `[[123]]`,
			html:     `<p><a href="/123/">123</a></p>`,
		},
		{
			markdown: `《[[123]]》`,
			html:     `<p>《<a href="/123/">123</a>》</p>`,
		},
		{
			// 代码优先级更高
			markdown: "`[[123]]`",
			html:     `<p><code>[[123]]</code></p>`,
		},
		{
			// 链接优先级更低
			markdown: "[123](123)",
			html:     `<p><a href="123">123</a></p>`,
		},
		{
			// 不存在/无权访问的文章
			markdown: `[[456]]`,
			html:     `<p><a href="/456/">页面不存在</a></p>`,
		},
		{
			markdown: `[[111]]`,
			html:     `<p><a href="/111/">&lt;p&gt;</a></p>`,
		},
	}
	for _, tc := range tests {
		md := renderers.NewMarkdown(page_link.New(context.Background(), getPageTitle, nil))
		html, err := md.Render(tc.markdown)
		if err != nil {
			t.Error(err)
			continue
		}
		if strings.TrimSpace(html) != tc.html {
			t.Errorf("不相等：\n%s\n%s", tc.html, html)
			continue
		}
	}
}
