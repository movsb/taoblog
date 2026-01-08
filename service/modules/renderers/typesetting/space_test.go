package typesetting_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/typesetting"
)

func TestSpace(t *testing.T) {
	for i, tc := range []struct {
		markdown string
		html     string
	}{
		{
			markdown: `a b`,
			html:     `<p>a<span class="space"> </span>b</p>`,
		},
		{
			markdown: `a  b`,
			html:     `<p>a<span class="space">  </span>b</p>`,
		},
	} {
		r := renderers.NewMarkdown(typesetting.NewSpace())
		output, err := r.Render(tc.markdown)
		if err != nil {
			t.Errorf(`render error: %v`, err)
			continue
		}
		if strings.TrimSpace(output) != tc.html {
			t.Errorf("not equal: #%d\nwant:\n%s\ngot:\n%s\n", i, tc.html, output)
			continue
		}
	}
}
