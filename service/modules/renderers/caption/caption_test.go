package caption_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/caption"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/setup/migration"
)

func TestRender(t *testing.T) {
	testCases := []struct {
		Markdown string
		HTML     string
	}{
		{
			Markdown: `![](a.avif)`,
			HTML:     `<p><img src="a.avif" alt=""/></p>`,
		},
		{
			Markdown: `![](1.avif)`,
			HTML:     `<figure><img src="1.avif" alt=""/><figcaption>普通&lt;文本&gt;说明</figcaption></figure>`,
		},
		{
			Markdown: `![](2.avif)`,
			HTML:     `<figure><img src="2.avif" alt=""/><figcaption><p><strong>bold</strong> <em>italic</em> 说明</p></figcaption></figure>`,
		},
	}

	db1 := migration.InitPosts(``, false)
	db2 := migration.InitFiles(``)
	dbFS := storage.NewSQLite(db1, storage.NewDataStore(db2))
	utils.Must(utils.Write(
		utils.Must1(dbFS.ForPost(1)),
		&proto.FileSpec{
			Path: `1.avif`,
			Mode: 0600,
			Size: 3,
			Time: 0,
			Meta: &proto.FileSpec_Meta{
				Source: &proto.FileSpec_Meta_Source{
					Format:  proto.FileSpec_Meta_Source_Plaintext,
					Caption: `普通<文本>说明`,
				},
			},
		},
		strings.NewReader("123")),
	)
	utils.Must(utils.Write(
		utils.Must1(dbFS.ForPost(1)),
		&proto.FileSpec{
			Path: `2.avif`,
			Mode: 0600,
			Size: 3,
			Time: 0,
			Meta: &proto.FileSpec_Meta{
				Source: &proto.FileSpec_Meta_Source{
					Format:  proto.FileSpec_Meta_Source_Markdown,
					Caption: `**bold** *italic* 说明`,
				},
			},
		},
		strings.NewReader("234")),
	)

	for i, tc := range testCases {
		cap := caption.New(gold_utils.NewWebFileSystem(
			utils.Must1(dbFS.ForPost(1)),
			utils.Must1(url.Parse(`/`)),
		))
		md := renderers.NewMarkdown(cap)
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
