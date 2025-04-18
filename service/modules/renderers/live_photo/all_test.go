package live_photo

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	test_utils "github.com/movsb/taoblog/modules/utils/test"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/media_size"
)

func TestRender(t *testing.T) {
	type Case struct {
		Markdown string
		HTML     string
	}
	cases := test_utils.MustLoadCasesFromYaml[Case](`test_data/tests.yaml`)
	dir := os.DirFS(`test_data`)
	web := gold_utils.NewWebFileSystem(dir, utils.Must1(url.Parse(`/`)))
	for _, tc := range cases {
		md := renderers.NewMarkdown(
			media_size.New(web),
			New(context.Background(), web),
		)
		html := utils.Must1(md.Render(tc.Markdown))
		if html != tc.HTML {
			t.Fatal(`not equal:`, "\n", tc.HTML, "\n", html)
		}
	}
}
