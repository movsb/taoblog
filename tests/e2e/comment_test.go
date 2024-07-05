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

// 评论的图片、链接的 scheme 不允许非法内容。
func TestCommentInvalidLinkScheme(t *testing.T) {
	contents := []string{
		`<javascript:alert(1);>`,
		`[](javascript:alert)`,
		`![](javascript:)`,
	}

	for _, content := range contents {
		rsp, err := client.CreateComment(guest,
			&proto.Comment{
				PostId:     1,
				Author:     `昵称`,
				Email:      fakeEmailAddress,
				SourceType: `markdown`,
				Source:     content,
			},
		)
		if err == nil {
			t.Errorf(`应该失败，但没有：%q`, content)
			continue
		}
		if !strings.Contains(err.Error(), `不支持的协议`) {
			t.Errorf(`未包含“不支持的协议”`)
		}
		_ = rsp
	}
}
