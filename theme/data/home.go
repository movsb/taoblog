package data

import (
	"context"
	"slices"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
)

type GroupedPosts struct {
	Name  string
	Posts []*Post
}

type HomeData struct {
	Tops     []*Post // 置顶文章列表。
	Shared   []*Post // 分享文章列表。
	Posts    []*Post
	Tweets   []*Post
	Comments []*LatestCommentsByPost

	Grouped []GroupedPosts

	PostCount    int64
	PageCount    int64
	CommentCount int64
}

func NewDataForHome(ctx context.Context, service proto.TaoBlogServer, impl service.ToBeImplementedByRpc) *Data {
	ac := auth.Context(ctx)
	// settings := utils.Must1(service.GetUserSettings(ctx, &proto.GetUserSettingsRequest{}))
	d := &Data{
		Context: ctx,
		svc:     service,
		User:    auth.Context(ctx).User,
	}

	home := &HomeData{
		PostCount:    impl.GetDefaultIntegerOption("post_count", 0),
		PageCount:    impl.GetDefaultIntegerOption("page_count", 0),
		CommentCount: impl.GetDefaultIntegerOption("comment_count", 0),
	}

	if !ac.User.IsGuest() {
		topPosts := utils.Must1(service.GetTopPosts(ctx,
			&proto.GetTopPostsRequest{
				GetPostOptions: &proto.GetPostOptions{
					WithLink:       proto.LinkKind_LinkKindRooted,
					ContentOptions: co.For(co.HomeLatestPosts),
				},
			})).Posts
		for _, p := range topPosts {
			pp := newPost(p)
			home.Tops = append(home.Tops, pp)
		}

		sharedPosts := utils.Must1(service.ListPosts(ctx,
			&proto.ListPostsRequest{
				Limit:     10,
				OrderBy:   "date DESC",
				Kinds:     []string{`post`, `tweet`},
				Ownership: proto.Ownership_OwnershipShared,
				GetPostOptions: &proto.GetPostOptions{
					WithLink:       proto.LinkKind_LinkKindRooted,
					ContentOptions: co.For(co.HomeLatestPosts),
				},
			})).Posts
		for _, p := range sharedPosts {
			pp := newPost(p)
			home.Shared = append(home.Shared, pp)
		}

		cats := utils.Must1(service.ListCategories(ctx, &proto.ListCategoriesRequest{}))
		cats.Categories = append(cats.Categories, &proto.Category{
			Id:   0,
			Name: "未分类",
		})
		for _, cat := range cats.Categories {
			posts := utils.Must1(service.ListPosts(ctx,
				&proto.ListPostsRequest{
					Limit:     10,
					OrderBy:   "date DESC",
					Ownership: proto.Ownership_OwnershipMine,
					GetPostOptions: &proto.GetPostOptions{
						WithLink:       proto.LinkKind_LinkKindRooted,
						ContentOptions: co.For(co.HomeLatestPosts),
					},
					Categories: []int32{cat.Id},
				},
			))
			if len(posts.Posts) == 0 {
				continue
			}
			group := GroupedPosts{
				Name: cat.Name,
			}
			for _, p := range posts.Posts {
				pp := newPost(p)
				group.Posts = append(group.Posts, pp)
			}
			home.Grouped = append(home.Grouped, group)
		}

		slices.SortFunc(home.Grouped, func(a, b GroupedPosts) int {
			return -int(a.Posts[0].Date - b.Posts[0].Date)
		})
	}

	ownership := utils.IIF(d.User.IsAdmin(), proto.Ownership_OwnershipAll, proto.Ownership_OwnershipMine)
	rsp, err := service.ListPosts(ctx,
		&proto.ListPostsRequest{
			Limit:     15,
			OrderBy:   "date DESC",
			Kinds:     []string{`post`},
			Ownership: ownership,
			GetPostOptions: &proto.GetPostOptions{
				WithLink:       proto.LinkKind_LinkKindRooted,
				ContentOptions: co.For(co.HomeLatestPosts),
			},
		},
	)
	if err != nil {
		panic(err)
	}
	for _, p := range rsp.Posts {
		pp := newPost(p)
		home.Posts = append(home.Posts, pp)
	}

	// 最近碎碎念
	{
		tweets, err := service.ListPosts(ctx,
			&proto.ListPostsRequest{
				Limit:     15,
				OrderBy:   `date desc`,
				Kinds:     []string{`tweet`},
				Ownership: ownership,
				GetPostOptions: &proto.GetPostOptions{
					WithLink:       proto.LinkKind_LinkKindRooted,
					ContentOptions: co.For(co.HomeLatestTweets),
				},
			},
		)
		if err != nil {
			panic(err)
		}
		for _, p := range tweets.Posts {
			pp := newPost(p)
			home.Tweets = append(home.Tweets, pp)
		}
	}

	comments, err := d.svc.ListComments(ctx,
		&proto.ListCommentsRequest{
			Types:   []string{},
			Limit:   15,
			OrderBy: "date DESC",

			Ownership:      proto.Ownership_OwnershipMineAndShared,
			ContentOptions: co.For(co.HomeLatestComments),
		})
	if err != nil {
		panic(err)
	}
	postsMap := make(map[int64]*LatestCommentsByPost)
	for _, c := range comments.Comments {
		p, ok := postsMap[c.PostId]
		if !ok {
			post, err := d.svc.GetPost(ctx,
				&proto.GetPostRequest{
					Id: int32(c.PostId),
					GetPostOptions: &proto.GetPostOptions{
						ContentOptions: co.For(co.HomeLatestCommentsPosts),
					},
				},
			)
			if err != nil {
				panic(err)
			}
			p = &LatestCommentsByPost{
				Post: newPost(post),
			}
			postsMap[c.PostId] = p
			home.Comments = append(home.Comments, p)
		}
		p.Comments = append(p.Comments, &Comment{
			Comment: c,
		})
	}

	d.Data = home
	return d
}
