package math_test

import (
	_ "embed"
	"strings"
	"sync"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	katex "github.com/movsb/taoblog/service/modules/renderers/math"
)

func TestServerMathMLConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			md := renderers.NewMarkdown(katex.New(), renderers.WithoutTransform())
			for _, source := range []string{`$x^2$`, `$\ce{CO2 + C -> 2 CO}$`, `$\frac{1}{2}$`} {
				html, err := md.Render(source)
				if err != nil || !strings.Contains(html, "<math") {
					t.Errorf("math output for %q: %s %v", source, html, err)
				}
			}
		}()
	}
	wg.Wait()
}

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
