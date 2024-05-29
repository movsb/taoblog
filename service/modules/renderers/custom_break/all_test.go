package custom_break_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers/custom_break"
	"github.com/yuin/goldmark"
)

func TestMarkdown(t *testing.T) {
	testCases := []struct {
		ID       float32
		Markdown string
		HTML     string
	}{
		{
			ID:       1.0,
			Markdown: `---`,
			HTML:     `<hr>`,
		},
		{
			ID:       2.0,
			Markdown: `--- 111 ---`,
			HTML:     `<div class="divider"><span>111</span></div>`,
		},
	}

	md := goldmark.New(goldmark.WithExtensions(custom_break.New()))

	for _, tc := range testCases {
		buf := bytes.NewBuffer(nil)
		if err := md.Convert([]byte(tc.Markdown), buf); err != nil {
			panic(err)
		}

		got := strings.TrimSpace(buf.String())
		want := strings.TrimSpace(tc.HTML)

		if got != want {
			t.Errorf("not equal:\nmarkdown:\n%s\nwant:\n%s\ngot:\n%s\n", tc.Markdown, want, got)
		}
	}
}
