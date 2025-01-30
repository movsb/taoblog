package e2e_test

import (
	"context"
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
	if p1.Id != 2 {
		panic(`应该=2`)
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

	eq(`管理员看别人分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户2看别人分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{})

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

	// var acl []*models.AccessControlEntry
	// r.server.DB().MustFind(&acl)
	// yaml.NewEncoder(os.Stdout).Encode(acl)

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p2.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p3.Id})

	eq(`管理员看别人分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{p1.Id})
	eq(`用户2看别人分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{p2.Id})

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

	eq(`管理员看别人分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{p1.Id})
	eq(`用户2看别人分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{p1.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p3.Id})
}
