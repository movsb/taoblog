package exif

import (
	"os"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/exif/exif_exports"
)

func TestOrientation(t *testing.T) {
	t.SkipNow()
	f := utils.Must1(os.Open(`test_data/rotate.avif`))
	defer f.Close()
	m := utils.Must1(exif_exports.Extract(f))
	if !m.SwapSizes() {
		t.Error(`应该旋转尺寸。`)
	}
}
