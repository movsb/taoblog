package data

import (
	"context"
	"fmt"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
)

type TweetsData struct {
	Tweets []*Post
	Count  int
}

func (t *TweetsData) Last(i int) bool {
	return i == t.Count-1
}

func NewDataForTweets(ctx context.Context, cfg *config.Config, svc protocols.TaoBlogServer) *Data {
	d := &Data{
		Meta: &MetaData{
			Title: fmt.Sprintf(`%s的碎碎念`, cfg.Comment.Author),
		},
		User:   auth.Context(ctx).User,
		Config: cfg,
		svc:    svc,
		Tweets: &TweetsData{},
	}

	posts, err := svc.ListPosts(ctx,
		&protocols.ListPostsRequest{
			Limit:    1000,
			Kinds:    []string{`tweet`},
			OrderBy:  `date desc`,
			WithLink: protocols.LinkKind_LinkKindRooted,
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
	for _, p := range posts.Posts {
		pp := newPost(p)
		d.Tweets.Tweets = append(d.Tweets.Tweets, pp)
	}
	d.Tweets.Count = len(d.Tweets.Tweets)

	return d
}
