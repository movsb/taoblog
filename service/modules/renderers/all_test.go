package renderers_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
)

func TestMarkdown(t *testing.T) {
	tr := renderers.NewMarkdown()
	t2, s, err := tr.Render(`### <a id="my-header"></a>Header`)
	fmt.Println(t2, s, err)
}

func TestImage(t *testing.T) {
	tr := renderers.NewMarkdown()
	t2, s, err := tr.Render(`
# heading

![a](a.png)
`)
	fmt.Println(t2, s, err)
}

func TestMarkdownAll(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("test_data")))
	defer server.Close()
	host := server.URL

	cases := []struct {
		ID          int
		Options     []renderers.Option
		Description string
		Markdown    string
		Html        string
	}{
		{
			ID:       1,
			Markdown: `![avatar](test_data/avatar.jpg?scale=0.3)`,
			Html:     `<p><img src="test_data/avatar.jpg" alt="avatar" loading="lazy" width=138 height=138 /></p>`,
		},
		{
			ID:       2,
			Markdown: `- ![avatar](test_data/avatar.jpg?scale=0.3)`,
			Html: `<ul>
<li><img src="test_data/avatar.jpg" alt="avatar" loading="lazy" width=138 height=138 /></li>
</ul>`,
		},
		{
			ID:          3,
			Description: `支持网络图片的缩放`,
			Markdown:    fmt.Sprintf(`![](%s/avatar.jpg?scale=0.1)`, host),
			Html:        fmt.Sprintf(`<p><img src="%s/avatar.jpg" alt="" loading="lazy" width=46 height=46 /></p>`, host),
		},
	}
	for _, tc := range cases {
		md := renderers.NewMarkdown(tc.Options...)
		if tc.ID == 2 {
			log.Println(`debug`)
		}
		_, html, err := md.Render(tc.Markdown)
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(html) != strings.TrimSpace(tc.Html) {
			t.Fatalf("not equal:\n%s\n%s\n%s\n\n", tc.Markdown, tc.Html, html)
		}
	}
}
