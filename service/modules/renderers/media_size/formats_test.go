package media_size

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	test_utils "github.com/movsb/taoblog/modules/utils/test"
)

func TestSvg(t *testing.T) {
	type _Svg struct {
		SVG    string `yaml:"svg"`
		Output string `yaml:"output"`
	}
	tcs := test_utils.MustLoadCasesFromYaml[_Svg]("test_data/svg.yaml")
	for i, tc := range tcs {
		md, err := svg(bytes.NewReader([]byte(tc.SVG)))
		if err != nil {
			t.Error(err)
			continue
		}
		output := fmt.Sprintf("%d %d", md.Width, md.Height)
		if output != tc.Output {
			t.Errorf(`svg case %d failed`, i)
			continue
		}
	}
}

func TestAvif(t *testing.T) {
	var testCases = []struct {
		path          string
		width, height int
	}{
		{
			path:   `test_data/test.avif`,
			width:  168,
			height: 76,
		},
	}

	for i, tc := range testCases {
		fp := utils.Must(os.Open(tc.path))
		defer fp.Close()

		md := utils.Must(avif(fp))
		if md.Width != tc.width || md.Height != tc.height {
			t.Errorf(`avif not equal: %d`, i)
		}
	}
}
