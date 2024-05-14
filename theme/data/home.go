package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// HomeData ...
type HomeData struct {
	Posts         []*Post
	PostComments  []*LatestCommentsByPost
	Tweets        []*Post
	TweetComments []*LatestCommentsByPost

	PostCount    int64
	PageCount    int64
	CommentCount int64
}

// NewDataForHome ...
func NewDataForHome(ctx context.Context, cfg *config.Config, service *service.Service) *Data {
	d := &Data{
		Config: cfg,
		User:   auth.Context(ctx).User,
		Meta:   &MetaData{},
		svc:    service,
	}
	home := &HomeData{
		PostCount:    service.GetDefaultIntegerOption("post_count", 0),
		PageCount:    service.GetDefaultIntegerOption("page_count", 0),
		CommentCount: service.GetDefaultIntegerOption("comment_count", 0),
	}
	posts := service.MustListPosts(ctx,
		&protocols.ListPostsRequest{
			Fields:  "id,title,type,status,date",
			Limit:   20,
			OrderBy: "date DESC",
		},
	)
	// 太 hardcode shit 了。
	for _, p := range posts {
		pp := newPost(p)
		pp.link = service.GetLink(p.Id)
		home.Posts = append(home.Posts, pp)
	}

	comments, err := d.svc.ListComments(ctx,
		&protocols.ListCommentsRequest{
			Types:   []string{`post`, `page`},
			Mode:    protocols.ListCommentsRequest_Flat,
			Limit:   15,
			OrderBy: "date DESC",

			ContentOptions: &protocols.PostContentOptions{
				PrettifyHtml: true,
			},
		})
	if err != nil {
		panic(err)
	}
	postsMap := make(map[int64]*LatestCommentsByPost)
	for _, c := range comments.Comments {
		p, ok := postsMap[c.PostId]
		if !ok {
			p = &LatestCommentsByPost{
				Post: newPost(d.svc.MustGetPost(ctx, c.PostId, &protocols.PostContentOptions{
					WithContent:       true,
					RenderCodeBlocks:  true,
					UseAbsolutePaths:  true,
					OpenLinksInNewTab: protocols.PostContentOptions_OpenLinkInNewTabKindAll,
				})),
			}
			postsMap[c.PostId] = p
			home.PostComments = append(home.PostComments, p)
		}
		p.Comments = append(p.Comments, &Comment{
			Comment: c,
		})
	}

	// 最近碎碎念
	{
		tweets := service.MustListLatestTweets(ctx, 15, &protocols.PostContentOptions{
			WithContent:  true,
			PrettifyHtml: true,
		})
		// 太 hardcode shit 了。
		for _, p := range tweets {
			pp := newPost(p)
			pp.link = service.GetLink(p.Id)
			home.Tweets = append(home.Tweets, pp)
		}
		comments, err := d.svc.ListComments(ctx,
			&protocols.ListCommentsRequest{
				Types:   []string{`tweet`},
				Mode:    protocols.ListCommentsRequest_Flat,
				Limit:   15,
				OrderBy: "date DESC",

				ContentOptions: &protocols.PostContentOptions{
					WithContent:  true,
					PrettifyHtml: true,
				},
			})
		if err != nil {
			panic(err)
		}
		postsMap := make(map[int64]*LatestCommentsByPost)
		for _, c := range comments.Comments {
			p, ok := postsMap[c.PostId]
			if !ok {
				p = &LatestCommentsByPost{
					Post: newPost(d.svc.MustGetPost(ctx, c.PostId, &protocols.PostContentOptions{
						WithContent:       true,
						RenderCodeBlocks:  true,
						UseAbsolutePaths:  true,
						OpenLinksInNewTab: protocols.PostContentOptions_OpenLinkInNewTabKindAll,
					})),
				}
				postsMap[c.PostId] = p
				home.TweetComments = append(home.TweetComments, p)
			}
			p.Comments = append(p.Comments, &Comment{Comment: c})
		}
	}

	d.Home = home
	return d
}
