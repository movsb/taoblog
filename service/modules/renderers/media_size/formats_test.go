package media_size

import (
	"bytes"
	"fmt"
	"image"
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
	tcs := test_utils.MustLoadCasesFromYaml[_Svg]("testdata/svg.yaml")
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
			path:   `testdata/test.avif`,
			width:  168,
			height: 76,
		},
	}

	for i, tc := range testCases {
		fp := utils.Must1(os.Open(tc.path))
		defer fp.Close()

		c, _ := utils.Must2(image.DecodeConfig(fp))
		if c.Width != tc.width || c.Height != tc.height {
			t.Errorf(`avif not equal: %d`, i)
		}
	}
}
