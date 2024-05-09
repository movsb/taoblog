package data

import (
	"context"
	"fmt"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
)

type TweetsData struct {
	Tweets []*Post
	Count  int
}

func (t *TweetsData) Last(i int) bool {
	return i == t.Count-1
}

func NewDataForTweets(ctx context.Context, cfg *config.Config, svc *service.Service) *Data {
	d := &Data{
		Meta: &MetaData{
			Title: fmt.Sprintf(`%s的碎碎念`, cfg.Comment.Author),
		},
		User:   auth.Context(ctx).User,
		Config: cfg,
		svc:    svc,
		Tweets: &TweetsData{},
	}

	posts := svc.MustListLatestTweets(ctx)
	for _, p := range posts {
		pp := newPost(p)
		pp.link = svc.GetPlainLink(p.Id)
		d.Tweets.Tweets = append(d.Tweets.Tweets, pp)
	}
	d.Tweets.Count = len(d.Tweets.Tweets)

	return d
}
