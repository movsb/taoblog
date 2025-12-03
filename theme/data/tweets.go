package data

import (
	"context"

	"github.com/movsb/taoblog/modules/utils"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
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
		User:    user.Context(ctx).User,
		svc:     svc,
	}

	d.Meta.Title = d.TweetName()

	user := user.Context(ctx).User
	ownership := utils.IIF(user.IsAdmin(), proto.Ownership_OwnershipAll, proto.Ownership_OwnershipMineAndShared)

	posts, err := svc.ListPosts(ctx,
		&proto.ListPostsRequest{
			Limit:     1000,
			Kinds:     []string{`tweet`},
			OrderBy:   `date desc`,
			Ownership: ownership,
			GetPostOptions: &proto.GetPostOptions{
				WithLink:       proto.LinkKind_LinkKindRooted,
				ContentOptions: co.For(co.Tweets),
			},
		},
	)
	if err != nil {
		panic(err)
	}
	tweets := &TweetsData{
		Name: TweetName,
	}
	for _, p := range posts.Posts {
		pp := newPost(p)
		tweets.Tweets = append(tweets.Tweets, pp)
	}
	tweets.Count = len(tweets.Tweets)
	d.Data = tweets

	return d
}
