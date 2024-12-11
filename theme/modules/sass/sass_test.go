package sass_test

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/movsb/taoblog/theme/modules/sass"
)

//go:embed test_data/fs
var root embed.FS

func TestCompileFS(t *testing.T) {
	fsdir, _ := fs.Sub(root, `test_data/fs`)
	css, err := sass.CompileFS(fsdir, `style.scss`)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(css)
}
