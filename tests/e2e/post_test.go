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
	"strings"
	"testing"
	"time"

	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := Serve(ctx)

	create := func(user context.Context, p *proto.Post) *proto.Post {
		return utils.Must1(r.client.Blog.CreatePost(user, p))
	}

	pa := create(r.admin, &proto.Post{Source: `# admin`, SourceType: `markdown`})
	p1 := create(r.user1, &proto.Post{Source: `# user1`, SourceType: `markdown`})
	p2 := create(r.user2, &proto.Post{Source: `# user2`, SourceType: `markdown`})
	if pa.Id != 1 {
		panic(`应该=1`)
	}

	eq := listPostsEq(r, t)

	eq(`管理员自己的`, r.admin, proto.Ownership_OwnershipMine, []int64{pa.Id})
	eq(`用户1自己的`, r.user1, proto.Ownership_OwnershipMine, []int64{p1.Id})
	eq(`用户2自己的`, r.user2, proto.Ownership_OwnershipMine, []int64{p2.Id})

	eq(`管理员看别人公开和分享的`, r.admin, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户1看别人公开和分享的`, r.user1, proto.Ownership_OwnershipTheir, []int64{})
	eq(`用户2看别人公开和分享的`, r.user2, proto.Ownership_OwnershipTheir, []int64{})

	eq(`管理员看自己的和分享的`, r.admin, proto.Ownership_OwnershipMineAndShared, []int64{pa.Id})
	eq(`用户1看自己的和分享的`, r.user1, proto.Ownership_OwnershipMineAndShared, []int64{p1.Id})
	eq(`用户2看自己的和分享的`, r.user2, proto.Ownership_OwnershipMineAndShared, []int64{p2.Id})

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

	eq(`管理员看所有自己有权限看的`, r.admin, proto.Ownership_OwnershipAll, []int64{pa.Id, p1.Id})
	eq(`用户1看所有自己有权限看的`, r.user1, proto.Ownership_OwnershipAll, []int64{pa.Id, p1.Id})
	eq(`用户2看所有自己有权限看的`, r.user2, proto.Ownership_OwnershipAll, []int64{p1.Id, p2.Id})
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

// TODO 测试即便在添加了凭证的情况下仍然只返回公开文章。
func TestRSS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fixed := time.FixedZone(`TEST`, 3600)

	r := Serve(ctx, server.WithTimezone(fixed), server.WithRSS(true))

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

		rssURL := fmt.Sprintf(`http://%s/rss`, r.server.HTTPAddr())
		req := utils.Must1(http.NewRequestWithContext(
			context.Background(),
			http.MethodGet, rssURL, nil))
		req.Header.Add(`Authorization`, `token `+auth.SystemToken())
		rsp := utils.Must1(http.DefaultClient.Do(req))
		if rsp.StatusCode != 200 {
			t.Fatal(`statusCode != 200`)
		}
		return rsp
	}

	{
		rsp := request(false)
		defer rsp.Body.Close()

		const expectedOutput = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
	<title>未命名</title>
	<link>http://localhost:2564</link>
	<description></description>
	<lastBuildDate>Thu, 27 Mar 2025 01:00:00 TEST</lastBuildDate>
	<item>
		<title>user1</title>
		<link>http://localhost:2564/1/</link>
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

		const expectedOutput = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
	<title>未命名</title>
	<link>http://localhost:2564</link>
	<description></description>
	<lastBuildDate>Thu, 27 Mar 2025 01:00:00 TEST</lastBuildDate>
	<item>
		<title>user1</title>
		<link>http://localhost:2564/1/</link>
		<pubDate>Thu, 27 Mar 2025 00:00:00 TEST</pubDate>
		<description><![CDATA[]]></description>
	</item>
	<item>
		<title>user2</title>
		<link>http://localhost:2564/2/</link>
		<pubDate>Thu, 27 Mar 2025 00:00:00 TEST</pubDate>
		<description><![CDATA[]]></description>
	</item>
	<item>
		<title>admin</title>
		<link>http://localhost:2564/3/</link>
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
	r := Serve(context.Background())
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
		panic(`user1 不可见？`)
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
		panic(`user2 可见？`)
	}
}
