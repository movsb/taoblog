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

const fakeEmailAddress = `fake@twofei.com`

func TestCreateComment(t *testing.T) {
	rsp2, err := client.CreateComment(guest, &proto.Comment{
		PostId:     1,
		Author:     `昵称`,
		Email:      fakeEmailAddress,
		SourceType: `markdown`,
		Source:     `<marquee style="max-width: 100px;">（🏃逃……</marquee>`,
	})
	if err == nil || !strings.Contains(err.Error(), `不能包含`) {
		t.Fatal(rsp2, err)
	}
}

func TestThrottler(t *testing.T) {
	Server.Service.TestEnableRequestThrottler(true)
	defer Server.Service.TestEnableRequestThrottler(false)

	first := true
	for i := 0; i < 2; i++ {
		rsp, err := client.CreateComment(guest,
			&proto.Comment{
				PostId:     1,
				Author:     `昵称`,
				Email:      fakeEmailAddress,
				SourceType: `markdown`,
				Source:     `1`,
			},
		)
		if first {
			if err != nil {
				t.Fatalf(`第一次不应该错`)
			}
			first = false
		} else {
			if err == nil {
				t.Fatalf(`第二次应该错`)
			}
			if !strings.Contains(err.Error(), `过于频繁`) {
				t.Fatalf(`错误内容不正确。`)
			}
		}
		_ = rsp
	}
}

