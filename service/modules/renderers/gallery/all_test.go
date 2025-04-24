package gallery

import (
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	test_utils "github.com/movsb/taoblog/modules/utils/test"
	"github.com/movsb/taoblog/service/modules/renderers"
)

func TestRender(t *testing.T) {
	type Case struct {
		Markdown string
		HTML     string
	}
	cases := test_utils.MustLoadCasesFromYaml[Case](`test_data/tests.yaml`)
	for i, tc := range cases {
		if i == 4 {
			i += 0
		}
		md := renderers.NewMarkdown(
			New(),
		)
		html := utils.Must1(md.Render(tc.Markdown))
		if html != tc.HTML {
			t.Fatalf("not equal: #%d\n%s\n\n%s", i, tc.HTML, html)
		}
	}
}
