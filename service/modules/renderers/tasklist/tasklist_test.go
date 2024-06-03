package task_list_test

import (
	"bytes"
	"strings"
	"testing"

	test_utils "github.com/movsb/taoblog/modules/utils/test"
	task_list "github.com/movsb/taoblog/service/modules/renderers/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func TestTaskList(t *testing.T) {
	type TC struct {
		ID       float32
		Markdown string
		HTML     string
	}

	tcs := test_utils.MustLoadCasesFromYaml[TC](`test.yaml`)

	for _, tc := range tcs {
		if tc.ID != 3.0 {
			// continue
		}
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
