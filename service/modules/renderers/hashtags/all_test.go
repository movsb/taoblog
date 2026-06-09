package hashtags_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/hashtags"
)

func TestHashTags(t *testing.T) {
	md := renderers.NewMarkdown(hashtags.New(func(tag string) string {
		return tag
	}, nil))

	for _, tc := range []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `A #A`,
			HTML:     `<p>A <span class="hashtag"><a href="A">#A</a></span></p>`,
		},
		{
			Markdown: `#A`,
			HTML:     ``,
		},
		{
			Markdown: `#A  #B #C`,
			HTML:     ``,
		},
		{
			Markdown: `A #L100`,
			HTML:     `<p>A <span class="hashtag">#L100</span></p>`,
		},
		{
			Markdown: `A #L100-L200`,
			HTML:     `<p>A <span class="hashtag">#L100-L200</span></p>`,
		},
	} {
		html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Error(err)
			continue
		}
		if strings.TrimSpace(html) != tc.HTML {
			t.Errorf("\n%s\n%s\n%s", tc.Markdown, tc.HTML, html)
			continue
		}
	}
}
