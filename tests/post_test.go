package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/cmd/server/throttler"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func listPostsEq(r *R, t *testing.T) func(p string, u context.Context, ownership proto.Ownership, list []int64) {
	return func(p string, u context.Context, ownership proto.Ownership, list []int64) {
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
}

func TestListPosts(t *testing.T) {
	r := Serve(t.Context())

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	pa := create(r.admin, &proto.Post{Source: `# admin`, SourceType: `markdown`, Status: models.PostStatusPrivate})
	p1 := create(r.user1, &proto.Post{Source: `# user1`, SourceType: `markdown`, Status: models.PostStatusPrivate})
	p2 := create(r.user2, &proto.Post{Source: `# user2`, SourceType: `markdown`, Status: models.PostStatusPrivate})
	p3 := create(r.user1, &proto.Post{Source: `# user1 & draft`, SourceType: `markdown`, Status: models.PostStatusDraft})
	if pa.Id != 1 {
		panic(`应该=1`)
	}

	eq := listPostsEq(r, t)

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{pa.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p2.Id})

	eq(`管理员自己的（含草稿）`, r.admin, proto.Ownership_OwnershipDrafts, []int64{})
	eq(`用户1自己的（含草稿）`, r.user1, proto.Ownership_OwnershipDrafts, []int64{p3.Id})
	eq(`用户2自己的（含草稿）`, r.user2, proto.Ownership_OwnershipDrafts, []int64{})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p2.Id})

	eq(`管理员看分享的`, r.admin, proto.Ownership_OwnershipShared, []int64{})
	eq(`用户1看分享的`, r.user1, proto.Ownership_OwnershipShared, []int64{})
	eq(`用户2看分享的`, r.user2, proto.Ownership_OwnershipShared, []int64{})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{pa.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{p1.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p2.Id})

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     pa.Id,
		Status: models.PostStatusPartial,
	}))
	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p1.Id,
		Status: models.PostStatusPartial,
	}))

	utils.Must1(r.client.Blog.SetPostACL(r.admin, &proto.SetPostACLRequest{
		PostId: pa.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user1ID): {
				Perms: []proto.Perm{
					proto.Perm_PermRead,
				},
			},
		},
	}))

	utils.Must1(r.client.Blog.SetPostACL(r.admin, &proto.SetPostACLRequest{
		PostId: p1.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user2ID): {
				Perms: []proto.Perm{
					proto.Perm_PermRead,
				},
			},
		},
	}))

	// 当前权限：
	//
	// pa → u1, p1 → u2, p2 公开

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{pa.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p2.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{pa.Id})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{p1.Id})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id, p1.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id, p2.Id})

	eq(`管理员看分享的`, r.admin, proto.Ownership_OwnershipShared, []int64{})
	eq(`用户1看分享的`, r.user1, proto.Ownership_OwnershipShared, []int64{pa.Id})
	eq(`用户2看分享的`, r.user2, proto.Ownership_OwnershipShared, []int64{p1.Id})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{pa.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{pa.Id, p1.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p1.Id,
		Status: models.PostStatusPrivate,
	}))

	// 当前权限：
	//
	// pa → u1, p1 → 私有, p2 → 公开

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{pa.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p2.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{pa.Id})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id, p1.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p2.Id})

	eq(`管理员看分享的`, r.admin, proto.Ownership_OwnershipShared, []int64{})
	eq(`用户1看分享的`, r.user1, proto.Ownership_OwnershipShared, []int64{pa.Id})
	eq(`用户2看分享的`, r.user2, proto.Ownership_OwnershipShared, []int64{})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{pa.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{pa.Id, p1.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p2.Id})

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p1.Id,
		Status: models.PostStatusPublic,
	}))

	// 当前权限：
	//
	// pa → u1, p1 → 公开, p2 → 公开

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{pa.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p2.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{p1.Id})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{pa.Id})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{p1.Id})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id, p1.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p2.Id})

	eq(`管理员看分享的`, r.admin, proto.Ownership_OwnershipShared, []int64{})
	eq(`用户1看分享的`, r.user1, proto.Ownership_OwnershipShared, []int64{pa.Id})
	eq(`用户2看分享的`, r.user2, proto.Ownership_OwnershipShared, []int64{})

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{pa.Id, p1.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{pa.Id, p1.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})
}

// 测试只可访问公开的文章。
func TestSitemaps(t *testing.T) {
	r := Serve(t.Context())

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	p1 := create(r.admin, &proto.Post{Status: models.PostStatusPublic, Source: `# admin`, SourceType: `markdown`})
	p2 := create(r.user1, &proto.Post{Status: models.PostStatusPrivate, Source: `# user1`, SourceType: `markdown`})
	log.Println(`状态：`, p1.Status, p2.Status)

	// TODO hard-coded URL
	rsp := utils.Must1(http.Get(r.server.JoinPath(`sitemap.xml`)))
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		t.Fatal(`状态码不正确。`)
	}
	buf := bytes.NewBuffer(nil)
	utils.Must1(io.Copy(buf, rsp.Body))
	// t.Log(buf.String())

	// NOTE: 硬编码的，难得解析了。
	expect := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url><loc>` + r.server.JoinPath(`/1/`) + `</loc></url>
</urlset>
`

	if buf.String() != expect {
		t.Fatalf("Sitemap 不匹配：\n%s\n%s", buf.String(), expect)
	}
}

func TestGetPost(t *testing.T) {
	r := Serve(t.Context())

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	p1 := create(r.user1, &proto.Post{Status: models.PostStatusPublic, Source: `# user1`, SourceType: `markdown`})
	p2 := create(r.user2, &proto.Post{Status: models.PostStatusPrivate, Source: `# user2`, SourceType: `markdown`})

	// 当前权限：p1 公开，p2 私有

	eq := func(p string, u context.Context, id int64, ok bool) {
		_, file, line, _ := runtime.Caller(1)
		_, err := r.client.Blog.GetPost(u, &proto.GetPostRequest{
			Id: int32(id),
		})
		if ok == (err == nil) {
			return
		}
		t.Errorf("[%s:%d]%s: %d, %v, %v", file, line, p, id, ok, err)
	}

	eq(`访客访问公开`, r.guest, p1.Id, true)
	eq(`访客访问私有`, r.guest, p2.Id, false)
	eq(`用户1自己的`, r.user1, p1.Id, true)
	eq(`用户2自己的`, r.user2, p2.Id, true)
	eq(`用户1访问用户2的私有`, r.user1, p2.Id, false)
	eq(`用户2访问用户1的公开`, r.user2, p1.Id, true)

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

	// 当前权限：p1 公开，p2 → u1

	eq(`访客访问分享`, r.guest, p2.Id, false)
	eq(`用户1访问用户2的分享`, r.user1, p2.Id, true)

	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p2.Id,
		Status: models.PostStatusPrivate,
	}))

	// 当前权限：p1 公开，p2 私有（曾经分享过给 u1，改为私有后不会删除分享记录）
	eq(`用户1访问用户2曾经分享过而现为私有`, r.user1, p2.Id, false)
}

// TODO 测试即便在添加了凭证的情况下仍然只返回公开文章。
func TestRSS(t *testing.T) {
	fixed := time.FixedZone(`TEST`, 3600)

	r := Serve(t.Context(), server.WithTimezone(fixed), server.WithRSS(true))

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	now := time.Date(2025, time.March, 27, 0, 0, 0, 0, fixed)

	p1 := create(r.user1, &proto.Post{Status: models.PostStatusPublic, Date: int32(now.Unix()), Source: `# user1`, SourceType: `markdown`})
	p2 := create(r.user2, &proto.Post{Status: models.PostStatusPrivate, Date: int32(now.Unix()), Source: `# user2`, SourceType: `markdown`})
	p3 := create(r.admin, &proto.Post{Status: models.PostStatusPartial, Date: int32(now.Unix()), Source: `# admin`, SourceType: `markdown`})
	_, _, _ = p1, p2, p3

	r.server.Main().TestingSetLastPostedAt(now.Add(time.Hour))

	request := func(pri bool) *http.Response {
		r.server.RSS().TestingEnablePrivate(pri)

		req := utils.Must1(http.NewRequestWithContext(
			context.Background(), http.MethodGet,
			r.server.JoinPath(`rss`), nil))
		r.addAuth(req, int64(user.SystemID))
		rsp := utils.Must1(http.DefaultClient.Do(req))
		if rsp.StatusCode != 200 {
			t.Fatal(`statusCode != 200`)
		}
		return rsp
	}

	{
		rsp := request(false)
		defer rsp.Body.Close()

		expectedOutput := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
	<title>未命名</title>
	<link>` + r.server.JoinPath() + `</link>
	<description></description>
	<lastBuildDate>Thu, 27 Mar 2025 01:00:00 TEST</lastBuildDate>
	<item>
		<title>user1</title>
		<link>` + r.server.JoinPath(`/1/`) + `</link>
		<pubDate>Thu, 27 Mar 2025 00:00:00 TEST</pubDate>
		<description><![CDATA[]]></description>
	</item>
</channel>
</rss>
`

		buf := bytes.NewBuffer(nil)
		io.Copy(buf, rsp.Body)

		if buf.String() != expectedOutput {
			t.Fatalf("RSS 输出不相等：\ngot:%s\nwant:%s\n", buf.String(), expectedOutput)
		}
	}

	{
		rsp := request(true)
		defer rsp.Body.Close()

		expectedOutput := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
	<title>未命名</title>
	<link>` + r.server.JoinPath() + `</link>
	<description></description>
	<lastBuildDate>Thu, 27 Mar 2025 01:00:00 TEST</lastBuildDate>
	<item>
		<title>user1</title>
		<link>` + r.server.JoinPath(`/1/`) + `</link>
		<pubDate>Thu, 27 Mar 2025 00:00:00 TEST</pubDate>
		<description><![CDATA[]]></description>
	</item>
	<item>
		<title>user2</title>
		<link>` + r.server.JoinPath(`/2/`) + `</link>
		<pubDate>Thu, 27 Mar 2025 00:00:00 TEST</pubDate>
		<description><![CDATA[]]></description>
	</item>
	<item>
		<title>admin</title>
		<link>` + r.server.JoinPath(`/3/`) + `</link>
		<pubDate>Thu, 27 Mar 2025 00:00:00 TEST</pubDate>
		<description><![CDATA[]]></description>
	</item>
</channel>
</rss>
`

		buf := bytes.NewBuffer(nil)
		io.Copy(buf, rsp.Body)

		if buf.String() != expectedOutput {
			t.Fatalf("RSS 输出不相等：\ngot:%s\nwant:%s\n", buf.String(), expectedOutput)
		}
	}
}

// 测试文章缓存权限的隔离。
// 案例：一篇公开的文章引用（通过page-link）了一篇私有文章。
//
//	那么，作者应该可见引用名，访客则不可见。
//
// TODO 评论区登录后没有动态刷新。
func TestIsolatedPostCache(t *testing.T) {
	r := Serve(t.Context())
	privatePost := utils.Must1(r.client.Blog.CreatePost(r.user1,
		&proto.Post{
			Status: models.PostStatusPrivate,
			Title:  `PRIVATE`,
			Source: `123`,
		},
	))
	publicPost := utils.Must1(r.client.Blog.CreatePost(r.user1,
		&proto.Post{
			Status: models.PostStatusPublic,
			Title:  `PUBLIC`,
			Source: fmt.Sprintf(`[[%d]]`, privatePost.Id),
		},
	))

	// TODO 应该 GET /ID/ 来判断，而不是 API，否则不够 E2E
	rendered := utils.Must1(r.client.Blog.GetPost(r.user1,
		&proto.GetPostRequest{
			Id: int32(publicPost.Id),
			GetPostOptions: &proto.GetPostOptions{
				ContentOptions: &proto.PostContentOptions{
					WithContent: true,
				},
			},
		},
	))
	if !strings.Contains(rendered.Content, `PRIVATE`) {
		panic(`user1 应该可见。`)
	}
	rendered = utils.Must1(r.client.Blog.GetPost(r.user2,
		&proto.GetPostRequest{
			Id: int32(publicPost.Id),
			GetPostOptions: &proto.GetPostOptions{
				ContentOptions: &proto.PostContentOptions{
					WithContent: true,
				},
			},
		},
	))
	if strings.Contains(rendered.Content, `PRIVATE`) {
		panic(`user2 不应该可见。`)
	}
}

func TestReferences(t *testing.T) {
	r := Serve(t.Context())
	privatePost := utils.Must1(r.client.Blog.CreatePost(r.user1,
		&proto.Post{
			Status: models.PostStatusPrivate,
			Title:  `PRIVATE`,
			Source: `123`,
		},
	))
	publicPost := utils.Must1(r.client.Blog.CreatePost(r.user1,
		&proto.Post{
			Status: models.PostStatusPublic,
			Title:  `PUBLIC`,
			Source: fmt.Sprintf(`[[%d]]`, privatePost.Id),
		},
	))
	publicPost2 := utils.Must1(r.client.Blog.CreatePost(r.user2,
		&proto.Post{
			Status: models.PostStatusPublic,
			Title:  `PUBLIC`,
			Source: fmt.Sprintf(`[[%d]][[%d]]`, publicPost.Id, privatePost.Id),
		},
	))

	gp := func(ctx context.Context, id int64) *proto.Post {
		p := utils.Must1(r.client.Blog.GetPost(ctx,
			&proto.GetPostRequest{Id: int32(id)},
		))
		if p.References == nil {
			p.References = &proto.Post_References{}
		}
		if p.References.Posts == nil {
			p.References.Posts = &proto.Post_References_Posts{}
		}
		return p
	}

	eq := func(a []int32, b []int64) {
		slices.Sort(a)
		slices.Sort(b)
		c := utils.Map(b, func(b int64) int32 { return int32(b) })
		if !reflect.DeepEqual(a, c) {
			if len(a) == 0 && len(c) == 0 {
				return
			}
			_, file, line, _ := runtime.Caller(1)
			log.Fatalf(`not equal: %s:%d %v %v`, file, line, a, c)
		}
	}

	p1 := gp(r.user1, privatePost.Id)
	eq(p1.References.Posts.From, []int64{publicPost.Id, publicPost2.Id})
	eq(p1.References.Posts.To, []int64{})
	p2 := gp(r.user1, publicPost.Id)
	eq(p2.References.Posts.From, []int64{publicPost2.Id})
	eq(p2.References.Posts.To, []int64{privatePost.Id})
	p3 := gp(r.user2, publicPost2.Id)
	eq(p3.References.Posts.From, []int64{})
	eq(p3.References.Posts.To, []int64{privatePost.Id, publicPost.Id})

	{
		utils.Must1(r.client.Blog.DeletePost(r.admin, &proto.DeletePostRequest{
			Id: int32(publicPost.Id),
		}))

		p1 := gp(r.user1, privatePost.Id)
		eq(p1.References.Posts.From, []int64{publicPost2.Id})
		eq(p1.References.Posts.To, []int64{})
		p3 := gp(r.user2, publicPost2.Id)
		eq(p3.References.Posts.From, []int64{})
		eq(p3.References.Posts.To, []int64{privatePost.Id})
	}
}

func TestUpdateTags(t *testing.T) {
	r := Serve(t.Context())
	p1 := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{Source: "# Tags\n#t1 #t2"}))
	p1 = utils.Must1(r.client.Blog.GetPost(r.user1, &proto.GetPostRequest{Id: int32(p1.Id), GetPostOptions: &proto.GetPostOptions{}}))
	tags := p1.Tags
	slices.Sort(tags)
	if !reflect.DeepEqual(tags, []string{`t1`, `t2`}) {
		t.Fatal(`not equal`)
	}
	p1.Source = "# Tags\n#t3"
	utils.Must1(r.client.Blog.UpdatePost(r.user1, &proto.UpdatePostRequest{
		Post: p1,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				`source_type`,
				`source`,
			},
		},
	}))
	p1 = utils.Must1(r.client.Blog.GetPost(r.user1, &proto.GetPostRequest{Id: int32(p1.Id), GetPostOptions: &proto.GetPostOptions{}}))
	tags = p1.Tags
	slices.Sort(tags)
	if !reflect.DeepEqual(tags, []string{`t3`}) {
		t.Fatal(`not equal`)
	}
}

func TestCreateUntitledPost(t *testing.T) {
	r := Serve(t.Context())
	_, err := r.client.Blog.CreateUntitledPost(r.guest, &proto.CreateUntitledPostRequest{Type: `markdown`})
	if err == nil {
		t.Fatal(`未鉴权`)
	}
	p, err := r.client.Blog.CreateUntitledPost(r.user1, &proto.CreateUntitledPostRequest{Type: `markdown`})
	if err != nil {
		t.Fatal(err)
	}
	if p.Post.Date != p.Post.Modified {
		t.Fatal(`创建的文章时间不正确`)
	}
	// 尝试修改文章，修改后“发表时间”应该是现在时间。
	// 暂时拿不到现在时间，所以 sleep 一下以确保不一样。
	time.Sleep(time.Second * 2)
	p2, err := r.client.Blog.UpdatePost(r.user1, &proto.UpdatePostRequest{
		Post:       p.Post,
		UpdateMask: &fieldmaskpb.FieldMask{},
	})
	if p2.Date == p.Post.Date {
		t.Fatal(`修改文章后，发表时间不应该等于最初的创建时间`)
	}
}

func TestUpdateTitle(t *testing.T) {
	r := Serve(t.Context())
	p1 := utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{Type: `tweet`, Source: "# Title"}))
	if p1.Title != `Title` {
		t.Fatal(`标题不正确`)
	}
	p1 = utils.Must1(r.client.Blog.UpdatePost(r.user1, &proto.UpdatePostRequest{
		Post: &proto.Post{
			Id:         p1.Id,
			Modified:   p1.Modified,
			SourceType: `markdown`,
			Source:     `no title`,
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{`source_type`, `source`},
		},
	}))
	// 文章原来的标题被删除，采用自动生成的。
	if p1.Title != `no title` {
		t.Fatal(`标题不正确`)
	}
}

func TestCreatePostThrottler(t *testing.T) {
	r := Serve(t.Context(),
		server.WithCreateFirstPost(),
		server.WithRequestThrottler(throttler.New()),
	)
	r.server.TestEnableRequestThrottler(true)
	defer r.server.TestEnableRequestThrottler(false)

	count := 1
	for range 4 {
		rsp, err := r.client.Blog.CreatePost(r.admin,
			&proto.Post{
				Source: `# Title`,
			},
		)
		switch count {
		case 1:
			if err != nil {
				t.Fatalf(`第一次不应该错`)
			}
		case 2:
			if err == nil {
				t.Fatalf(`第二次应该错`)
			}
			if !strings.Contains(err.Error(), `过于频繁`) {
				t.Fatalf(`错误内容不正确。`)
			}
			time.Sleep(time.Second * 3)
		case 3:
			if err != nil {
				t.Fatalf(`第三次应该不错误。`)
			}
		case 4:
			if err == nil {
				t.Fatalf(`第四次应该错`)
			}
		}
		count++
		_ = rsp
	}
}
