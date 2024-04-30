package data

import (
	"context"
	"fmt"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
)

// 马斯克应该不会告我吧？
type Tweet struct {
	*Post
}

type TweetsData struct {
	Tweets []*Tweet
	Count  int
}

func (t *TweetsData) Last(i int) bool {
	return i == t.Count-1
}

func NewDataForTweets(cfg *config.Config, user *auth.User, svc *service.Service) *Data {
	d := &Data{
		Meta: &MetaData{
			Title: fmt.Sprintf(`%s的碎碎念`, cfg.Comment.Author),
		},
		User:   user,
		Config: cfg,
		svc:    svc,
		Tweets: &TweetsData{},
	}

	posts := svc.MustListLatestTweets(user.Context(context.TODO()))
	for _, p := range posts {
		pp := newPost(p)
		pp.link = svc.GetPlainLink(p.Id)
		d.Tweets.Tweets = append(d.Tweets.Tweets, &Tweet{pp})
	}
	d.Tweets.Count = len(d.Tweets.Tweets)

	return d
}
