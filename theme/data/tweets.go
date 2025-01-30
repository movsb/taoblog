package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type TweetsData struct {
	Name   string
	Tweets []*Post
	Count  int
}

const TweetName = `碎碎念`

func NewDataForTweets(ctx context.Context, cfg *config.Config, svc proto.TaoBlogServer) *Data {
	d := &Data{
		ctx:    ctx,
		Meta:   &MetaData{},
		Config: cfg,
		svc:    svc,
		Tweets: &TweetsData{
			Name: TweetName,
		},
	}

	d.Meta.Title = d.TweetName()

	posts, err := svc.ListPosts(ctx,
		&proto.ListPostsRequest{
			Limit:          1000,
			Kinds:          []string{`tweet`},
			OrderBy:        `date desc`,
			WithLink:       proto.LinkKind_LinkKindRooted,
			ContentOptions: co.For(co.Tweets),
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
