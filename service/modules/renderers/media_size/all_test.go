package media_size_test

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/media_size"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `<iframe width=560 height=315></iframe>`,
			HTML:     `<iframe width="560" height="315" style="aspect-ratio:16/9;" class="landscape  too-wide"></iframe>`,
		},
	}

	for i, tc := range testCases {
		ext := media_size.New(
			gold_utils.NewWebFileSystem(os.DirFS(`test_data`), utils.Must1(url.Parse(`/`))),
			media_size.WithDimensionLimiter(350),
		)
		md := renderers.NewMarkdown(ext)
		html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Errorf("%d: %v", i, err.Error())
			continue
		}
		if strings.TrimSpace(html) != strings.TrimSpace(tc.HTML) {
			t.Errorf("%d not equal:\nmarkdown:\n%s\nwant:\n%s\ngot:\n%s\n", i, tc.Markdown, tc.HTML, html)
			continue
		}
	}
}
