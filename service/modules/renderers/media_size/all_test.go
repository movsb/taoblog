package media_size_test

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/media_size"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `<iframe width=560 height=315></iframe>`,
			HTML:     `<iframe width="560" height="315" style="aspect-ratio: 1.777778;"></iframe>`,
		},
	}

	for i, tc := range testCases {
		ext := media_size.New(
			gold_utils.NewWebFileSystem(os.DirFS(`testdata`), utils.Must1(url.Parse(`/`))),
		)
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

func TestArgs(t *testing.T) {
	testCases := []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `![](test.avif)`,
			HTML:     `<p><img src="test.avif" alt="" width="168" height="76"/></p>`,
		},
		{
			Markdown: `![](test.avif?scale=2)`,
			HTML:     `<p><img src="test.avif" alt="" width="336" height="152"/></p>`,
		},
		{
			Markdown: `![](test.avif?s=.5)`,
			HTML:     `<p><img src="test.avif" alt="" width="84" height="38"/></p>`,
		},
		{
			Markdown: `![](test.avif?cover)`,
			HTML:     `<p><img src="test.avif" alt="" style="object-fit: cover; aspect-ratio: 1; width: 400px" class="cover" width="168" height="76"/></p>`,
		},
		{
			Markdown: `![](test.avif?w=100)`,
			HTML:     `<p><img src="test.avif" alt="" width="100" height="45"/></p>`,
		},
	}

	for i, tc := range testCases {
		ext := media_size.New(
			gold_utils.NewWebFileSystem(os.DirFS(`testdata`), utils.Must1(url.Parse(`/`))),
		)
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
