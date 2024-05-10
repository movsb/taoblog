package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// HomeData ...
type HomeData struct {
	Posts    []*Post
	Comments []*LatestCommentsByPost

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
			Types: utils.IIF(
				auth.Context(ctx).User.IsAdmin(),
				nil, // TODO 允许管理员显示全部评论，暂时放这儿
				[]string{`post`, `page`},
			),

			Mode:    protocols.ListCommentsRequest_Flat,
			Limit:   15,
			OrderBy: "date DESC",

			PrettifyHtml: true,
		})
	if err != nil {
		panic(err)
	}
	postsMap := make(map[int64]*LatestCommentsByPost)
	for _, c := range comments.Comments {
		p, ok := postsMap[c.PostId]
		if !ok {
			p = &LatestCommentsByPost{
				PostID:    c.PostId,
				PostTitle: d.svc.GetPostTitle(c.PostId),
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
