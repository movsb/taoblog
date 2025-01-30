package e2e_test

import (
	"context"
	"strings"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

func TestPreviewComment(t *testing.T) {
	r := Serve(context.Background())
	rsp, err := r.client.Blog.PreviewComment(r.guest, &proto.PreviewCommentRequest{
		Markdown: `<a>`,
		PostId:   1,
	})
	if err == nil || !strings.Contains(err.Error(), "不能包含") {
		t.Fatal(rsp, err)
	}
}

const fakeEmailAddress = `fake@twofei.com`

func TestCreateComment(t *testing.T) {
	r := Serve(context.Background())
	rsp2, err := r.client.Blog.CreateComment(r.guest, &proto.Comment{
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
	r := Serve(context.Background())
	r.server.TestEnableRequestThrottler(true)
	defer r.server.TestEnableRequestThrottler(false)

	first := true
	for i := 0; i < 2; i++ {
		rsp, err := r.client.Blog.CreateComment(r.guest,
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
	r := Serve(context.Background())
	contents := []string{
		`<javascript:alert(1);>`,
		`[](javascript:alert)`,
		`![](javascript:)`,
	}

	for _, content := range contents {
		rsp, err := r.client.Blog.CreateComment(r.guest,
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

// 测试递归删除评论。
func TestDeleteCommentsRecursively(t *testing.T) {
	r := Serve(context.Background())
	post := utils.Must1(r.client.Blog.CreatePost(r.admin, &proto.Post{
		Type:       `post`,
		SourceType: `markdown`,
		Source:     "# 测试递归删除评论",
	}))

	/*
		c1
			c1.1
				c1.1.1
			c1.2
				c1.2.1
	*/
	create := func(parent int64) *proto.Comment {
		return utils.Must1(r.client.Blog.CreateComment(r.admin, &proto.Comment{
			PostId:     post.Id,
			Parent:     parent,
			Author:     `author`,
			Email:      fakeEmailAddress,
			SourceType: `markdown`,
			Source:     `test`,
		}))
	}
	c1 := create(0)
	c11 := create(c1.Id)
	c111 := create(c11.Id)
	_ = c111

	c12 := create(c1.Id)
	c121 := create(c12.Id)
	_ = c121

	count := utils.Must1(r.client.Blog.GetPostCommentsCount(r.admin, &proto.GetPostCommentsCountRequest{PostId: post.Id})).Count
	if count != 5 {
		t.Fatalf(`评论数应该为 5 条。`)
	}

	// 删除 c11 后应该剩 3 条。
	utils.Must1(r.client.Blog.DeleteComment(r.admin, &proto.DeleteCommentRequest{Id: int32(c11.Id)}))

	count2 := utils.Must1(r.client.Blog.GetPostCommentsCount(r.admin, &proto.GetPostCommentsCountRequest{PostId: post.Id})).Count
	if count2 != 3 {
		t.Fatalf(`评论数应该为 3 条，但是剩余：%d 条`, count2)
	}
}
