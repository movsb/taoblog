package image_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/image"
)

func TestRenderByType(t *testing.T) {
	m := renderers.NewMarkdown(image.New())

	testCases := []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `![](1.jpg)`,
			HTML:     `<p><img src="1.jpg" alt=""></p>`,
		},
		{
			Markdown: `![](1.mp3)`,
			HTML:     `<p><audio controls src="1.mp3"></audio></p>`,
		},
		{
			Markdown: `![](1.mp4)`,
			HTML:     `<p><video controls src="1.mp4"></video></p>`,
		},
	}

	for _, tc := range testCases {
		output, err := m.Render(tc.Markdown)
		if err != nil {
			t.Errorf(`错误：%s`, tc.Markdown)
			continue
		}
		if strings.TrimSpace(output) != tc.HTML {
			t.Errorf(`结果不一样：%s, %s`, output, tc.HTML)
			continue
		}
	}
}
