package renderers_test

import (
	"log"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
)

func TestPrettifier(t *testing.T) {
	cases := []struct {
		ID          float32
		Options     []renderers.Option2
		Description string
		Markdown    string
		Text        string
	}{
		{
			ID:       1,
			Markdown: `![文本](/URL)`,
			Text:     `[图片:文本]`,
		},
		{
			ID:       2,
			Markdown: `[文本](链接)`,
			Text:     `[链接:文本]`,
		},
		{
			ID:       3,
			Markdown: "代码：\n\n```go\npackage main\n```",
			Text:     "代码：\n[代码]",
		},
		{
			ID: 4,
			Markdown: `|h1|h2|
|-|-|
|1|2|
`,
			Text: `[表格]`,
		},
	}
	for _, tc := range cases {
		md := renderers.NewMarkdown(tc.Options...)
		if tc.ID == 6.0 {
			log.Println(`debug`)
		}
		_, html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Fatal(err)
		}
		text, err := (&renderers.Prettifier{}).Prettify(html)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(text) != strings.TrimSpace(tc.Text) {
			t.Fatalf("not equal:\nMarkdown:%s\nExpected:%s\nGot:%s\n\n", tc.Markdown, tc.Text, text)
		}
	}
}
