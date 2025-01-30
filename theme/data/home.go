package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
)

// HomeData ...
type HomeData struct {
	Posts    []*Post
	Tweets   []*Post
	Comments []*LatestCommentsByPost

	PostCount    int64
	PageCount    int64
	CommentCount int64
}

// NewDataForHome ...
func NewDataForHome(ctx context.Context, cfg *config.Config, service proto.TaoBlogServer, impl service.ToBeImplementedByRpc) *Data {
	d := &Data{
		ctx:    ctx,
		Config: cfg,
		Meta:   &MetaData{},
		svc:    service,
	}
	home := &HomeData{
		PostCount:    impl.GetDefaultIntegerOption("post_count", 0),
		PageCount:    impl.GetDefaultIntegerOption("page_count", 0),
		CommentCount: impl.GetDefaultIntegerOption("comment_count", 0),
	}
	rsp, err := service.ListPosts(ctx,
		&proto.ListPostsRequest{
			Limit:          15,
			OrderBy:        "date DESC",
			Kinds:          []string{`post`},
			WithLink:       proto.LinkKind_LinkKindRooted,
			ContentOptions: co.For(co.HomeLatestPosts),
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
				Limit:          15,
				OrderBy:        `date desc`,
				Kinds:          []string{`tweet`},
				WithLink:       proto.LinkKind_LinkKindRooted,
				ContentOptions: co.For(co.HomeLatestTweets),
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
			Mode:    proto.ListCommentsRequest_Flat,
			Limit:   15,
			OrderBy: "date DESC",

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
					Id:             int32(c.PostId),
					ContentOptions: co.For(co.HomeLatestCommentsPosts),
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

	d.Home = home
	return d
}
