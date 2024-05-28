package task_list_test

import (
	"bytes"
	"strings"
	"testing"

	task_list "github.com/movsb/taoblog/service/modules/renderers/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func TestTaskList(t *testing.T) {
	testCases := []struct {
		ID       float32
		Markdown string
		HTML     string
	}{
		{
			ID:       1.0,
			Markdown: "- [ ] Task1\n- [ ] Task2",
			HTML: `<ul>
<li><input type=checkbox disabled> Task1</li>
<li><input type=checkbox disabled> Task2</li>
</ul>
`,
		},
		{
			ID:       2.0,
			Markdown: "<!-- TaskList -->\n\n- [ ] Task1\n- [x] Task2",
			HTML: `<!-- TaskList -->
<ul class="task-list">
<li class="task-list-item" data-source-position="22"><input type=checkbox disabled autocomplete="off"> Task1</li>
<li class="checked task-list-item" data-source-position="34"><input type=checkbox checked disabled autocomplete="off"> Task2</li>
</ul>
`,
		},
	}

	for _, tc := range testCases {
		md := goldmark.New(
			goldmark.WithExtensions(task_list.New()),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		)
		buf := bytes.NewBuffer(nil)
		if err := md.Convert([]byte(tc.Markdown), buf); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if strings.TrimSpace(out) != strings.TrimSpace(tc.HTML) {
			t.Errorf("Mismatch: id=%v\n\nMarkdown:\n%s\nWant:\n|%s|\nGot:\n|%s|\n\n",
				tc.ID, tc.Markdown, tc.HTML, out)
		}
	}
}
