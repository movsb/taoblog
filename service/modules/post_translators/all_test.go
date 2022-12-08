package post_translators

import (
	"fmt"
	"testing"
)

func TestMarkdown(t *testing.T) {
	tr := &MarkdownTranslator{}
	t2, s, err := tr.Translate(`### <a id="my-header"></a>Header`)
	fmt.Println(t2, s, err)
}

func TestImage(t *testing.T) {
	tr := &MarkdownTranslator{}
	t2, s, err := tr.Translate(`
# heading

![a](a.png)
`)
	fmt.Println(t2, s, err)
}
