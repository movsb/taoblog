package e2e_test

import (
	"context"
	"strings"
	"testing"

	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/cmd/server/throttler"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPreviewComment(t *testing.T) {
	r := Serve(t.Context(), server.WithCreateFirstPost())
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
	r := Serve(t.Context(), server.WithCreateFirstPost())
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

func TestCreateCommentRequiresReadablePost(t *testing.T) {
	r := Serve(t.Context())

	publicPost := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		Source:     `# public`,
		SourceType: `markdown`,
		Status:     models.PostStatusPublic,
	}))
	privatePost := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		Source:     `# private`,
		SourceType: `markdown`,
		Status:     models.PostStatusPrivate,
	}))
	draftPost := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		Source:     `# draft`,
		SourceType: `markdown`,
		Status:     models.PostStatusDraft,
	}))
	partialPost := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		Source:     `# partial`,
		SourceType: `markdown`,
		Status:     models.PostStatusPartial,
	}))

	comment := func(postID int64) *proto.Comment {
		return &proto.Comment{
			PostId:     postID,
			Author:     `昵称`,
			Email:      fakeEmailAddress,
			SourceType: `markdown`,
			Source:     `test`,
		}
	}
	expectNotFound := func(name string, postID int64, ctx context.Context) {
		t.Helper()
		_, err := r.client.Blog.CreateComment(ctx, comment(postID))
		if status.Code(err) != codes.NotFound {
			t.Fatalf(`%s: got %v, want NotFound`, name, err)
		}
	}

	utils.Must1(r.client.Blog.CreateComment(r.guest, comment(publicPost.Id)))
	utils.Must1(r.client.Blog.CreateComment(r.user1, comment(privatePost.Id)))
	utils.Must1(r.client.Blog.CreateComment(r.user1, comment(draftPost.Id)))

	expectNotFound(`guest private`, privatePost.Id, r.guest)
	expectNotFound(`guest draft`, draftPost.Id, r.guest)
	expectNotFound(`guest partial`, partialPost.Id, r.guest)
	expectNotFound(`other private`, privatePost.Id, r.user2)
	expectNotFound(`other draft`, draftPost.Id, r.user2)
	expectNotFound(`other partial`, partialPost.Id, r.user2)
	expectNotFound(`missing post`, 999999, r.guest)

	utils.Must1(r.client.Blog.SetPostACL(r.admin, &proto.SetPostACLRequest{
		PostId: partialPost.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user2ID): {
				Perms: []proto.Perm{proto.Perm_PermRead},
			},
		},
	}))
	utils.Must1(r.client.Blog.CreateComment(r.user2, comment(partialPost.Id)))
}

func TestThrottler(t *testing.T) {
	r := Serve(t.Context(),
		server.WithCreateFirstPost(),
		server.WithRequestThrottler(throttler.New()),
	)
	r.server.TestEnableRequestThrottler(true)
	defer r.server.TestEnableRequestThrottler(false)

	first := true
	for range 2 {
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
	r := Serve(t.Context(), server.WithCreateFirstPost())
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
			t.Errorf(`未包含“不支持的协议”：%v`, err.Error())
		}
		_ = rsp
	}
}

// 测试递归删除评论。
func TestDeleteCommentsRecursively(t *testing.T) {
	r := Serve(t.Context())
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
