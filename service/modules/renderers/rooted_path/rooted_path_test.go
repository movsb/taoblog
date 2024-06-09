package rooted_path_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
	"github.com/movsb/taoblog/service/modules/renderers/rooted_path"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		Base     string
		Markdown string
		HTML     string
	}{
		{
			Base:     `/123`,
			Markdown: `<a href="1.avif"></a>`,
			HTML:     `<p><a href="/1.avif"></a></p>`,
		},
		{
			Base:     `https://blog.twofei.com/123/`,
			Markdown: `<a href="1.avif?a=b"></a>`,
			HTML:     `<p><a href="/123/1.avif?a=b"></a></p>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<a href="1.avif"></a>`,
			HTML:     `<p><a href="/123/1.avif"></a></p>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<img src="1.avif"/>`,
			HTML:     `<img src="/123/1.avif"/>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<iframe src="1.html"></iframe>`,
			HTML:     `<iframe src="/123/1.html"></iframe>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<source src="1.mp3"/>`,
			HTML:     `<source src="/123/1.mp3"/>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<audio src="1.mp3"></audio>`,
			HTML:     `<p><audio src="/123/1.mp3"></audio></p>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<video src="1.mp4"></video>`,
			HTML:     `<p><video src="/123/1.mp4"></video></p>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<object data="1.pdf"></object>`,
			HTML:     `<p><object data="/123/1.pdf"></object></p>`,
		},
		{
			Base:     `/123/`,
			Markdown: `<object data="中文.pdf"></object>`,
			HTML:     `<p><object data="/123/%E4%B8%AD%E6%96%87.pdf"></object></p>`,
		},
	}

	for i, tc := range testCases {
		ext := rooted_path.New(gold_utils.NewWebFileSystem(nil, utils.Must(url.Parse(tc.Base))))
		md := renderers.NewMarkdown(ext)
		html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Errorf("%d: %v", i, err.Error())
			continue
		}
		if strings.TrimSpace(html) != strings.TrimSpace(tc.HTML) {
			t.Errorf("%d not equal:\nmarkdown:\n%s\nwant:\n%s\ngot:\n%s\n", i, tc.Markdown, tc.HTML, html)
			continue
		}
	}
}
