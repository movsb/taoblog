package data

import (
	"context"
	"fmt"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

type TweetsData struct {
	Tweets []*Post
	Count  int
}

func (t *TweetsData) Last(i int) bool {
	return i == t.Count-1
}

func NewDataForTweets(ctx context.Context, cfg *config.Config, svc protocols.TaoBlogServer, impl service.ToBeImplementedByRpc) *Data {
	d := &Data{
		Meta: &MetaData{
			Title: fmt.Sprintf(`%s的碎碎念`, cfg.Comment.Author),
		},
		User:   auth.Context(ctx).User,
		Config: cfg,
		svc:    svc,
		impl:   impl,
		Tweets: &TweetsData{},
	}

	posts, err := svc.ListPosts(ctx,
		&protocols.ListPostsRequest{
			Limit:   1000,
			Kinds:   []string{`tweet`},
			OrderBy: `date desc`,
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
		pp.link = impl.GetPlainLink(p.Id)
		d.Tweets.Tweets = append(d.Tweets.Tweets, pp)
	}
	d.Tweets.Count = len(d.Tweets.Tweets)

	return d
}
