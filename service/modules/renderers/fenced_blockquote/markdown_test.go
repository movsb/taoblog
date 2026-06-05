package fenced_blockquote_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers/fenced_blockquote"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func TestMarkdown(t *testing.T) {
	testCases := []struct {
		name     string
		markdown string
		html     string
	}{
		{
			name: "paragraph",
			markdown: `"""
hello
"""`,
			html: `<blockquote>
<p>hello</p>
</blockquote>`,
		},
		{
			name: "blocks",
			markdown: `"""
hello

` + "```go" + `
fmt.Println("hello")
` + "```" + `

- a
- b
"""`,
			html: `<blockquote>
<p>hello</p>
<pre><code class="language-go">fmt.Println(&quot;hello&quot;)
</code></pre>
<ul>
<li>a</li>
<li>b</li>
</ul>
</blockquote>`,
		},
		{
			name: "nested",
			markdown: `""""
outer

"""
inner
"""

outer again
""""`,
			html: `<blockquote>
<p>outer</p>
<blockquote>
<p>inner</p>
</blockquote>
<p>outer again</p>
</blockquote>`,
		},
		{
			name: "ordinary text",
			markdown: `""" quote
text`,
			html: `<p>&quot;&quot;&quot; quote
text</p>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			md := goldmark.New(goldmark.WithExtensions(extension.GFM, fenced_blockquote.New()))
			buf := bytes.NewBuffer(nil)
			if err := md.Convert([]byte(tc.markdown), buf); err != nil {
				t.Fatal(err)
			}
			got := strings.TrimSpace(buf.String())
			want := strings.TrimSpace(tc.html)
			if got != want {
				t.Errorf("not equal:\nmarkdown:\n%s\nwant:\n%s\ngot:\n%s\n", tc.markdown, want, got)
			}
		})
	}
}
