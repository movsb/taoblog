package custom_break_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers/custom_break"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func TestMarkdown(t *testing.T) {
	testCases := []struct {
		ID       float32
		Markdown string
		Options  []goldmark.Option
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
		{
			ID: 3.0,
			Markdown: `时间|记录
---|---
2024-06-01|iOS
`,
			Options: []goldmark.Option{goldmark.WithExtensions(extension.GFM)},
			HTML: `
<table>
<thead>
<tr>
<th>时间</th>
<th>记录</th>
</tr>
</thead>
<tbody>
<tr>
<td>2024-06-01</td>
<td>iOS</td>
</tr>
</tbody>
</table>
`,
		},
	}

	for _, tc := range testCases {
		var options = []goldmark.Option{
			goldmark.WithExtensions(
				custom_break.New(),
			),
		}
		options = append(options, tc.Options...)
		md := goldmark.New(options...)

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
