package post_translators

import (
	"fmt"
	"testing"
)

func TestMarkdown(t *testing.T) {
	tr := &MarkdownTranslator{}
	s, err := tr.Translate(nil, `### <a id="my-header"></a>Header`, `/`)
	fmt.Println(s, err)
}
