package page_link_test

import (
	"context"
	"fmt"
	"reflect"
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
		refs     []int32
	}{
		{
			markdown: `[[123]]`,
			html:     `<p><a href="/123/">123</a></p>`,
			refs:     []int32{123},
		},
		{
			markdown: `《[[123]]》`,
			html:     `<p>《<a href="/123/">123</a>》</p>`,
			refs:     []int32{123},
		},
		{
			// 代码优先级更高
			markdown: "`[[123]]`",
			html:     `<p><code>[[123]]</code></p>`,
			refs:     []int32{},
		},
		{
			// 链接优先级更低
			markdown: "[123](123)",
			html:     `<p><a href="123">123</a></p>`,
			refs:     []int32{},
		},
		{
			// 不存在/无权访问的文章
			markdown: `[[456]]`,
			html:     `<p><a href="/456/">页面不存在</a></p>`,
			refs:     []int32{456},
		},
		{
			markdown: `[[111]]`,
			html:     `<p><a href="/111/">&lt;p&gt;</a></p>`,
			refs:     []int32{111},
		},
		{
			markdown: `[AAA](/123)`,
			html:     `<p><a href="/123">AAA</a></p>`,
			refs:     []int32{123},
		},
		{
			markdown: `[AAA](/123/)`,
			html:     `<p><a href="/123/">AAA</a></p>`,
			refs:     []int32{123},
		},
	}
	for i, tc := range tests {
		refs := []int32{}
		md := renderers.NewMarkdown(page_link.New(context.Background(), getPageTitle, &refs))
		html, err := md.Render(tc.markdown)
		if err != nil {
			t.Error(err)
			continue
		}
		if strings.TrimSpace(html) != tc.html || !reflect.DeepEqual(refs, tc.refs) {
			t.Errorf("不相等(%d)：\n%s\n%s\n%v %v", i, tc.html, html, tc.refs, refs)
			continue
		}
	}
}
