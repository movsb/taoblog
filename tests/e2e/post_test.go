package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"slices"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
)

func TestListPosts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := Serve(ctx)

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	p1 := create(r.admin, &proto.Post{Source: `# admin`, SourceType: `markdown`})
	p2 := create(r.user1, &proto.Post{Source: `# user1`, SourceType: `markdown`})
	p3 := create(r.user2, &proto.Post{Source: `# user2`, SourceType: `markdown`})
	if p1.Id != 1 {
		panic(`应该=1`)
	}

	eq := func(p string, u context.Context, ownership proto.Ownership, list []int64) {
		_, file, line, _ := runtime.Caller(1)
		ps := utils.Must1(r.client.Blog.ListPosts(u, &proto.ListPostsRequest{
			Ownership: ownership,
		})).Posts
		slices.SortFunc(ps, func(a, b *proto.Post) int { return int(a.Id - b.Id) })
		if slices.CompareFunc(ps, list, func(p *proto.Post, i int64) int {
			return int(p.Id - i)
		}) != 0 {
			t.Fatalf(`[%s:%d] %s: got: %v, expect: %v`, file, line, p, utils.Map(ps, func(p *proto.Post) int64 { return p.Id }), list)
		}
	}

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p2.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p3.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{p2.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p3.Id})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{p1.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{p2.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p3.Id})

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p1.Id,
		Status: models.PostStatusPartial,
	}))
	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p2.Id,
		Status: models.PostStatusPartial,
	}))

	utils.Must1(r.client.Blog.SetPostACL(r.admin, &proto.SetPostACLRequest{
		PostId: p1.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user1ID): {
				Perms: []proto.Perm{
					proto.Perm_PermRead,
				},
			},
		},
	}))

	utils.Must1(r.client.Blog.SetPostACL(r.admin, &proto.SetPostACLRequest{
		PostId: p2.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user2ID): {
				Perms: []proto.Perm{
					proto.Perm_PermRead,
				},
			},
		},
	}))

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p2.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p3.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{p1.Id})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{p2.Id})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id, p2.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p2.Id, p3.Id})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{p1.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p2.Id, p3.Id})

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p2.Id,
		Status: models.PostStatusPrivate,
	}))

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p2.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p3.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{p1.Id})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id, p2.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p3.Id})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{p1.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p3.Id})

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p2.Id,
		Status: models.PostStatusPublic,
	}))

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p2.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p3.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{p2.Id})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{p1.Id})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{p2.Id})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id, p2.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p3.Id})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p2.Id, p3.Id})
}

// 测试只可访问公开的文章。
func TestSitemaps(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := Serve(ctx)

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	p1 := create(r.admin, &proto.Post{Status: models.PostStatusPublic, Source: `# admin`, SourceType: `markdown`})
	p2 := create(r.user1, &proto.Post{Status: models.PostStatusPrivate, Source: `# user1`, SourceType: `markdown`})
	log.Println(`状态：`, p1.Status, p2.Status)

	// TODO hard-coded URL
	u := fmt.Sprintf(`http://%s/sitemap.xml`, r.server.HTTPAddr())
	rsp := utils.Must1(http.Get(u))
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		t.Fatal(`状态码不正确。`)
	}
	buf := bytes.NewBuffer(nil)
	utils.Must1(io.Copy(buf, rsp.Body))
	// t.Log(buf.String())

	// TODO 硬编码的，难得解析了。
	expect := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url><loc>http://localhost:2564/1/</loc></url>
</urlset>
`

	if buf.String() != expect {
		t.Fatalf("Sitemap 不匹配：\n%s\n%s", buf.String(), expect)
	}
}

func TestGetPost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := Serve(ctx)

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	p1 := create(r.user1, &proto.Post{Status: models.PostStatusPublic, Source: `# user1`, SourceType: `markdown`})
	p2 := create(r.user2, &proto.Post{Status: models.PostStatusPrivate, Source: `# user2`, SourceType: `markdown`})

	eq := func(p string, u context.Context, id int32, ok bool) {
		_, file, line, _ := runtime.Caller(1)
		_, err := r.client.Blog.GetPost(u, &proto.GetPostRequest{
			Id: id,
		})
		if ok == (err == nil) {
			return
		}
		t.Errorf("[%s:%d]%s: %d, %v, %v", file, line, p, id, ok, err)
	}

	eq(`访客访问公开`, r.guest, int32(p1.Id), true)
	eq(`访客访问私有`, r.guest, int32(p2.Id), false)
	eq(`用户1自己的`, r.user1, int32(p1.Id), true)
	eq(`用户2自己的`, r.user2, int32(p2.Id), true)
	eq(`用户1访问用户2的私有`, r.user1, int32(p2.Id), false)
	eq(`用户2访问用户1的公开`, r.user2, int32(p1.Id), true)

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p2.Id,
		Status: models.PostStatusPartial,
	}))

	utils.Must1(r.client.Blog.SetPostACL(r.admin, &proto.SetPostACLRequest{
		PostId: p2.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user1ID): {
				Perms: []proto.Perm{
					proto.Perm_PermRead,
				},
			},
		},
	}))

	eq(`访客访问公开`, r.guest, int32(p1.Id), true)
	eq(`访客访问分享`, r.guest, int32(p2.Id), false)
	eq(`用户1自己的`, r.user1, int32(p1.Id), true)
	eq(`用户2自己的`, r.user2, int32(p2.Id), true)
	eq(`用户1访问用户2的分享`, r.user1, int32(p2.Id), true)
	eq(`用户2访问用户1的公开`, r.user2, int32(p2.Id), true)
}
