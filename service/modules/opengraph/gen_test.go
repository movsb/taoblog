package open_graph

import (
	"bytes"
	"os"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestGen(t *testing.T) {
	t.SkipNow()
	png, err := GenerateImage(
		`陪她去流浪`,
		`静候大作。像我现在都没有心情和时间写博客了，但我就喜欢看你们写。`,
		bytes.NewReader(utils.Must1(os.ReadFile(`av.png`))),
		bytes.NewReader(utils.Must1(os.ReadFile(`bg.avif`))),
	)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(`output.png`, png, 0644)
}
