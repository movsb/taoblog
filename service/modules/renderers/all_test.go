package renderers

import (
	"fmt"
	"testing"
)

func TestMarkdown(t *testing.T) {
	tr := &_Markdown{}
	t2, s, err := tr.Render(`### <a id="my-header"></a>Header`)
	fmt.Println(t2, s, err)
}

func TestImage(t *testing.T) {
	tr := &_Markdown{}
	t2, s, err := tr.Render(`
# heading

![a](a.png)
`)
	fmt.Println(t2, s, err)
}
