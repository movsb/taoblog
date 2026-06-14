package data

import (
	"context"
	"slices"

	"github.com/movsb/taoblog/modules/utils"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
)

type TweetsData struct {
	Name   string
	Tweets []*Post
	Count  int

	HasPrev bool
	HasNext bool

	NewestTime int
	OldestTime int
}

const TweetName = `碎碎念`

func NewDataForTweets(ctx context.Context, svc proto.TaoBlogServer, before, after int) *Data {
	d := &Data{
		Context: ctx,
		User:    user.Context(ctx).User,
		svc:     svc,
	}
	tweets := &TweetsData{
		Name: TweetName,
	}
	d.Meta.Title = d.TweetName()

	user := user.Context(ctx).User
	ownership := utils.IIF(user.IsAdmin(), proto.Ownership_OwnershipAll, proto.Ownership_OwnershipMineAndShared)

	req := &proto.ListPostsRequest{
		Limit:     30,
		Kinds:     []string{`tweet`},
		Ownership: ownership,
		GetPostOptions: &proto.GetPostOptions{
			WithLink:       proto.LinkKind_LinkKindRooted,
			ContentOptions: co.For(co.Tweets),
		},
	}

	// before 指往前翻，由于是倒序，所以是查更新的。
	// 两个不同时出现，所以用 else。
	if before > 0 {
		req.CreatedNotBefore = int32(before) + 1
		req.OrderBy = `date asc`
	} else if after > 0 {
		req.CreatedNotAfter = int32(after)
		req.OrderBy = `date desc`
	} else {
		req.OrderBy = `date desc`
	}

	posts, err := svc.ListPosts(ctx, req)
	if err != nil {
		panic(err)
	}
	// 如果是往新的翻，则是date asc排序的，需要逆序一下。
	if before > 0 {
		slices.Reverse(posts.Posts)
	}
	for _, p := range posts.Posts {
		pp := newPost(p)
		tweets.Tweets = append(tweets.Tweets, pp)
	}
	tweets.Count = len(tweets.Tweets)

	// 通过拉取前/后面的至少1篇文章来判断是否有上一页/下一页。
	// BUG：创建时间的精度是1秒，所以文章的更新时间如果在同一秒内发生，
	// 这里的逻辑会有问题。最好是在创建文章那里限制一下。
	// 目前不太可能出现这种现象，可以不考虑修复。
	if tweets.Count > 0 {
		newest := posts.GetPosts()[0]
		oldest := posts.GetPosts()[len(posts.GetPosts())-1]

		prevPosts := utils.Must1(svc.ListPosts(ctx,
			&proto.ListPostsRequest{
				Limit:            1,
				Kinds:            []string{`tweet`},
				Ownership:        ownership,
				CreatedNotBefore: newest.Date + 1,
			}),
		)
		nextPosts := utils.Must1(svc.ListPosts(ctx,
			&proto.ListPostsRequest{
				Limit:           1,
				Kinds:           []string{`tweet`},
				Ownership:       ownership,
				CreatedNotAfter: oldest.Date - 1,
			}),
		)

		tweets.HasPrev = len(prevPosts.Posts) > 0
		tweets.HasNext = len(nextPosts.Posts) > 0

		tweets.NewestTime = int(newest.Date)
		tweets.OldestTime = int(oldest.Date)
	}

	d.Data = tweets
	return d
}
