package auto_image_border

import (
	"image/color"
	"log"
	"os"
	"path"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestMinimal(t *testing.T) {
	for i, tc := range []struct {
		file       string
		background color.NRGBA
		good       bool
	}{
		{
			file:       `black.png`,
			background: color.NRGBA{255, 255, 255, 0},
			good:       true,
		},
		{
			file:       `black.png`,
			background: color.NRGBA{0, 0, 0, 0},
			good:       true,
		},
		{
			file:       `white.png`,
			background: color.NRGBA{255, 255, 255, 0},
			good:       false,
		},
		{
			file:       `white.png`,
			background: color.NRGBA{0, 0, 0, 0},
			good:       true,
		},
		{
			file:       `white.webp`,
			background: color.NRGBA{0, 0, 0, 0},
			good:       true,
		},
		{
			file:       `black.avif`,
			background: color.NRGBA{255, 255, 255, 0},
			good:       true,
		},
	} {
		f := utils.Must1(os.Open(path.Join(`testdata`, tc.file)))
		defer f.Close()

		value := BorderContrastRatio(f, tc.background.R, tc.background.G, tc.background.B, 1)
		if value > 0.75 != tc.good {
			t.Fatal(`not good:`, i+1, tc.file)
		}
	}
}

func TestFile(t *testing.T) {
	t.SkipNow()
	f := utils.Must1(os.Open(path.Join(`testdata`, `IMG_1303.avif`)))
	defer f.Close()
	value := BorderContrastRatio(f, 255, 255, 255, 1)
	log.Println(value)
}
