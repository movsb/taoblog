package open_graph

import (
	"bytes"
	"os"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestGen(t *testing.T) {
	t.SkipNow()
	// debug = true
	png, err := GenerateImage(
		`陪她去流浪`,
		`以前为了学习编程语言标准库，单独建，后来发现更新得很慢`,
		`这这里是一段摘要内容这里是一段摘要内容这里是一段摘要内容这里是一段摘要内容这里是一段摘要内容这里是一段摘要内容这里是一段摘要内容这里是一段摘要内容里是一段摘要内容`,
		bytes.NewReader(utils.Must1(os.ReadFile(`av.png`))),
		bytes.NewReader(utils.Must1(os.ReadFile(`bg.avif`))),
	)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(`output.png`, png, 0644)
}
