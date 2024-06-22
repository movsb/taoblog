package emojis_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/emojis"
)

func TestEmojis(t *testing.T) {
	testCases := []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `[旺柴]`,
			HTML:     `<p><img class="emoji weixin" title="旺柴" src="/v3/dynamic/assets/weixin/doge.png"/></p>`,
		},
		{
			Markdown: `[旺柴][旺柴]`,
			HTML:     `<p><img class="emoji weixin" title="旺柴" src="/v3/dynamic/assets/weixin/doge.png"/><img class="emoji weixin" title="旺柴" src="/v3/dynamic/assets/weixin/doge.png"/></p>`,
		},
	}
	for _, tc := range testCases {
		md := renderers.NewMarkdown(emojis.New())
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
