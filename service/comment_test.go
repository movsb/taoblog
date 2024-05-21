package service_test

import (
	"strings"
	"testing"

	"github.com/movsb/taoblog/protocols"
)

func TestPreviewComment(t *testing.T) {
	initService()
	rsp, err := blog.PreviewComment(guest, &proto.PreviewCommentRequest{
		Markdown: `<a>`,
		PostId:   1,
	})
	if err == nil || !strings.Contains(err.Error(), "不能包含") {
		t.Fatal(rsp, err)
	}
}
