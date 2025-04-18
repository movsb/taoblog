package media_tags

import (
	"embed"
	"io/fs"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	test_utils "github.com/movsb/taoblog/modules/utils/test"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
)

//go:embed test_data/player.html
var player embed.FS

func TestRender(t *testing.T) {

	_gTmpl = utils.NewTemplateLoader(utils.Must1(fs.Sub(player, `test_data`)), nil, nil)

	type _TestCase struct {
		Markdown string `yaml:"markdown"`
		HTML     string `yaml:"html"`
	}

	testCases := test_utils.MustLoadCasesFromYaml[_TestCase](`test_data/tests.yaml`)

	for i, tc := range testCases {
		utils.ReInitTestRandomSeed()
		tag := New(gold_utils.NewWebFileSystem(
			os.DirFS("test_data"),
			utils.Must1(url.Parse(`/`)),
		))
		md := renderers.NewMarkdown(tag)
		html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Errorf("%d: %v", i, err.Error())
			continue
		}
		if strings.TrimSpace(html) != strings.TrimSpace(tc.HTML) {
			t.Errorf("%d not equal:\nmarkdown:\n%s\nwant:\n%s\ngot:\n%s\n", i, tc.Markdown, tc.HTML, html)
			// ioutil.WriteFile("1.html", []byte(tc.HTML), 0644)
			// ioutil.WriteFile("2.html", []byte(html), 0644)
			continue
		}
	}
}
