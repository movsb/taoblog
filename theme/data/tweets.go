package data

import (
	"context"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type TweetsData struct {
	Name   string
	Tweets []*Post
	Count  int
}

const TweetName = `碎碎念`

func NewDataForTweets(ctx context.Context, svc proto.TaoBlogServer) *Data {
	d := &Data{
		Context: ctx,
		svc:     svc,
		Tweets: &TweetsData{
			Name: TweetName,
		},
	}

	d.Meta.Title = d.TweetName()

	user := auth.Context(ctx).User
	ownership := utils.IIF(user.IsAdmin(), proto.Ownership_OwnershipAll, proto.Ownership_OwnershipMineAndShared)

	posts, err := svc.ListPosts(ctx,
		&proto.ListPostsRequest{
			Limit:          1000,
			Kinds:          []string{`tweet`},
			OrderBy:        `date desc`,
			WithLink:       proto.LinkKind_LinkKindRooted,
			ContentOptions: co.For(co.Tweets),
			Ownership:      ownership,
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
