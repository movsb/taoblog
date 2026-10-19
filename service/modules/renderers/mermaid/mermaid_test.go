package mermaid

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/movsb/taoblog/service/modules/renderers"
)

func TestServerRender(t *testing.T) {
	for _, source := range []string{
		"flowchart LR\n A[中文测试] --> B[成功]",
		"sequenceDiagram\n Alice->>Bob: 你好\n Bob-->>Alice: 收到",
		"classDiagram\n Animal <|-- Duck",
	} {
		svg, err := render(source)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(svg, "<svg ") {
			t.Fatal("missing SVG")
		}
		d := xml.NewDecoder(strings.NewReader(svg))
		for {
			_, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		if strings.Contains(source, "中文测试") && !strings.Contains(svg, "中文测试") {
			t.Fatal("Chinese label lost")
		}
	}
}

func TestMarkdown(t *testing.T) {
	md := renderers.NewMarkdown(renderers.WithFencedCodeBlockRenderer("mermaid", New()))
	result, err := md.Render("```mermaid\nflowchart LR\n A[开始] --> B[结束]\n```\n\n后续正文")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `<div class="mermaid"><svg`) || !strings.Contains(result, "后续正文") {
		t.Fatal(result)
	}
	result, err = md.Render("```mermaid\n<script>alert(1)</script>\n```\n\n后续正文")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result, "<script>") || !strings.Contains(result, "后续正文") {
		t.Fatal(result)
	}
	// A failed diagram must not poison subsequent renders.
	if _, err := render("flowchart LR\n A --> B"); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentDiagrams(t *testing.T) {
	var wg sync.WaitGroup
	ids := sync.Map{}
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out bytes.Buffer
			if err := New().RenderFencedCodeBlock(&out, "mermaid", nil, []byte("flowchart LR\n A --> B")); err != nil {
				t.Error(err)
				return
			}
			d := xml.NewDecoder(strings.NewReader(out.String()))
			for {
				token, err := d.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Error(err)
					return
				}
				if start, ok := token.(xml.StartElement); ok {
					for _, attr := range start.Attr {
						if attr.Name.Local == "id" {
							if _, loaded := ids.LoadOrStore(attr.Value, true); loaded {
								t.Errorf("duplicate SVG id: %s", attr.Value)
							}
						}
					}
				}
			}
		}()
	}
	wg.Wait()
}

func TestDisabledRejectsMarkdown(t *testing.T) {
	md := renderers.NewMarkdown(renderers.WithFencedCodeBlockRenderer("mermaid", New(WithDisabled(true))))
	for _, source := range []string{
		"```mermaid\nflowchart LR\n A --> B\n```",
		"```mermaid\n```",
		"```mermaid\ninvalid diagram\n```",
	} {
		_, err := md.Render(source)
		if err == nil || !strings.Contains(err.Error(), "评论中暂不允许") {
			t.Fatalf("disabled Mermaid must reject comment: %v", err)
		}
	}
	if _, err := md.Render("普通评论\n\n```go\nfmt.Println(1)\n```"); err != nil {
		t.Fatalf("ordinary comments must still render: %v", err)
	}
}
