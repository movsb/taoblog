package blur_image

import (
	"os"
	"testing"

	"github.com/movsb/taoblog/service/models"
)

func TestCalcHashAVIF(t *testing.T) {
	// 与线上 avifdec 路径生成的 ThumbHash 对照。
	for _, tc := range []struct {
		file string
		want string
	}{
		{"../../../../gateway/handlers/avatar/avatar.avif", "NQgGBwAImIdyiniHeWaHqHiIBwAAAAAA"},
		{"../exif/testdata/rotate.avif", "ThgKFYQMdopmh3hzmph3dmB3Bmdo"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			fp, err := os.Open(tc.file)
			if err != nil {
				t.Fatal(err)
			}
			got, err := calcHash(&models.File{Path: tc.file}, fp)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("ThumbHash = %q, want %q", got, tc.want)
			}
		})
	}
}
