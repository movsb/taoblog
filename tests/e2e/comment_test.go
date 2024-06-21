package e2e_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/protocols/go/proto"
)

func TestPreviewComment(t *testing.T) {
	rsp, err := client.PreviewComment(guest, &proto.PreviewCommentRequest{
		Markdown: `<a>`,
		PostId:   1,
	})
	if err == nil || !strings.Contains(err.Error(), "不能包含") {
		t.Fatal(rsp, err)
	}
}

func TestCreateComment(t *testing.T) {
	rsp2, err := client.CreateComment(guest, &proto.Comment{
		PostId:     1,
		Author:     `昵称`,
		Email:      `fake@twofei.com`,
		SourceType: `markdown`,
		Source:     `<marquee style="max-width: 100px;">（🏃逃……</marquee>`,
	})
	if err == nil || !strings.Contains(err.Error(), `不能包含`) {
		t.Fatal(rsp2, err)
	}
}
