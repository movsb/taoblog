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
func NewDataForHome(ctx context.Context, cfg *config.Config, service protocols.TaoBlogServer, impl service.ToBeImplementedByRpc) *Data {
	d := &Data{
		Config: cfg,
		User:   auth.Context(ctx).User,
		Meta:   &MetaData{},
		svc:    service,
	}
	home := &HomeData{
		PostCount:    impl.GetDefaultIntegerOption("post_count", 0),
		PageCount:    impl.GetDefaultIntegerOption("page_count", 0),
		CommentCount: impl.GetDefaultIntegerOption("comment_count", 0),
	}
	rsp, err := service.ListPosts(ctx,
		&protocols.ListPostsRequest{
			Limit:    20,
			OrderBy:  "date DESC",
			Kinds:    []string{`post`},
			WithLink: protocols.LinkKind_LinkKindRooted,
		},
	)
	if err != nil {
		panic(err)
	}
	// 太 hardcode shit 了。
	for _, p := range rsp.Posts {
		pp := newPost(p)
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
			post, err := d.svc.GetPost(ctx,
				&protocols.GetPostRequest{
					Id: int32(c.PostId),
					ContentOptions: &protocols.PostContentOptions{
						WithContent:       true,
						RenderCodeBlocks:  true,
						UseAbsolutePaths:  true,
						OpenLinksInNewTab: protocols.PostContentOptions_OpenLinkInNewTabKindAll,
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
			home.PostComments = append(home.PostComments, p)
		}
		p.Comments = append(p.Comments, &Comment{
			Comment: c,
		})
	}

	// 最近碎碎念
	{
		tweets, err := service.ListPosts(ctx,
			&protocols.ListPostsRequest{
				Limit:    15,
				OrderBy:  `date desc`,
				Kinds:    []string{`tweet`},
				WithLink: protocols.LinkKind_LinkKindRooted,
				ContentOptions: &protocols.PostContentOptions{
					WithContent:  true,
					PrettifyHtml: true,
				},
			},
		)
		if err != nil {
			panic(err)
		}
		// 太 hardcode shit 了。
		for _, p := range tweets.Posts {
			pp := newPost(p)
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
				post, err := d.svc.GetPost(ctx,
					&protocols.GetPostRequest{
						Id: int32(c.PostId),
						ContentOptions: &protocols.PostContentOptions{
							WithContent:       true,
							RenderCodeBlocks:  true,
							UseAbsolutePaths:  true,
							OpenLinksInNewTab: protocols.PostContentOptions_OpenLinkInNewTabKindAll,
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
				home.TweetComments = append(home.TweetComments, p)
			}
			p.Comments = append(p.Comments, &Comment{Comment: c})
		}
	}

	d.Home = home
	return d
}
