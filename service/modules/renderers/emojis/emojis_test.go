package emojis

import (
	"net/url"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
)

func TestEmojis(t *testing.T) {
	initEmojis()
	testCases := []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `[旺柴]`,
			HTML:     `<p><img src="/v3/dynamic/emojis/weixin/doge.png" alt="[旺柴]" title="旺柴" class="static emoji weixin"/></p>`,
		},
		{
			Markdown: `[未知]`,
			HTML:     `<p>[未知]</p>`,
		},
		{
			Markdown: `[旺柴][Doge]`,
			HTML:     `<p><img src="/v3/dynamic/emojis/weixin/doge.png" alt="[旺柴]" title="旺柴" class="static emoji weixin"/><img src="/v3/dynamic/emojis/weixin/doge.png" alt="[Doge]" title="Doge" class="static emoji weixin"/></p>`,
		},
	}
	baseURL, _ := url.Parse(`/v3/dynamic/emojis/`)
	for _, tc := range testCases {
		md := renderers.NewMarkdown(New(baseURL))
		output, err := md.Render(tc.Markdown)
		if err != nil {
			t.Error(err)
			continue
		}
		if strings.TrimSpace(output) != strings.TrimSpace(tc.HTML) {
			t.Errorf("not equal:\nMarkdown:%s\nWanted:%s\nGot   :%s\n",
				tc.Markdown, tc.HTML, output)
			continue
		}
	}
}
