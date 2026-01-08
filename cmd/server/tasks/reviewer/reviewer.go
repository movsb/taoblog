package reviewer

import (
	"context"
	"fmt"
	"iter"
	"log"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/calendar/solar"
)

var reviewerCalKind = calendar.RegisterKind(func(e *calendar.Event) string {
	return fmt.Sprint(e.PostID)
})

func RunReviewersTask(ctx context.Context, svc *service.Service) {
	t := &_ReviewerTask{s: svc}
	t.Run(ctx)
}

type _ReviewerTask struct {
	s *service.Service
}

func (t *_ReviewerTask) Run(ctx context.Context) {
	time.Sleep(time.Second * 10)
	t.run(ctx)
	utils.AtMiddleNight(ctx, func() { t.run(ctx) })
}

func (t *_ReviewerTask) run(ctx context.Context) {
	listRsp, err := t.s.ListPosts(
		user.SystemForLocal(ctx),
		&proto.ListPostsRequest{
			GetPostOptions: &proto.GetPostOptions{
				ContentOptions: &proto.PostContentOptions{},
				WithLink:       proto.LinkKind_LinkKindFull,
			},
		},
	)
	if err != nil {
		log.Println(`列举文章列表失败：`, err)
		return
	}

	t.s.CalenderService().Remove(reviewerCalKind, func(e *calendar.Event) bool {
		return true
	})

	userEnabled := map[int]bool{}

	for _, p := range listRsp.GetPosts() {
		if p.Status == models.PostStatusDraft {
			continue
		}
		if p.Type == `page` {
			continue
		}

		enabled, ok := userEnabled[int(p.UserId)]
		if !ok {
			u, err := t.s.GetUserSettings(
				user.SystemForLocal(ctx),
				&proto.GetUserSettingsRequest{
					UserId: uint32(p.UserId),
				},
			)
			if err != nil {
				log.Println(err)
				continue
			}
			userEnabled[int(p.UserId)] = u.ReviewPostsInCalendar
			enabled = u.ReviewPostsInCalendar
		}
		if !enabled {
			continue
		}

		t.scheduleNext(ctx, p)
	}

	// 1: 不重要，排除 hello world 后。
	if len(listRsp.GetPosts()) > 1 {
		log.Println(`Reviewer Task: Done.`)
	}
}

// 为文章 p 安排下次审阅时间。
//
// NOTE: Post 是不包含 Content 的（即没有包含最终 HTML）。
//
// TODO: 没有处理文章时区信息。
func (t *_ReviewerTask) scheduleNext(ctx context.Context, p *proto.Post) {
	var (
		now, _       = solar.AllDay(time.Now())
		createdAt, _ = solar.AllDay(time.Unix(int64(p.Date), 0).Local())
	)
	var (
		last time.Time
		next time.Time
	)

	for tt := range yieldTimes(createdAt) {
		if !tt.Before(now) {
			next = tt
			break
		}
		last = tt
	}

	// 上次事件，只保留近期 N 天的。
	latest := func(a, b time.Time, n int) bool {
		return b.Sub(a) <= time.Hour*24*time.Duration(n)
	}

	// lazy evaluation.
	render := sync.OnceValue(func() string {
		pp, err := t.s.GetPost(user.SystemForLocal(ctx), &proto.GetPostRequest{
			Id: int32(p.Id),
			GetPostOptions: &proto.GetPostOptions{
				ContentOptions: &proto.PostContentOptions{
					WithContent:  true,
					PrettifyHtml: true,
				},
			},
		})
		if err != nil {
			log.Println(`渲染出错：`, err, p.Id)
			return ``
		}
		return pp.Content
	})

	for tt := range func(yield func(time.Time) bool) {
		if !last.IsZero() && latest(last, now, 3) {
			if !yield(last) {
				return
			}
		}
		if !next.IsZero() && latest(now, next, 7) {
			if !yield(next) {
				return
			}
		}
	} {
		st, et := solar.AllDay(tt)

		t.s.CalenderService().AddEvent(reviewerCalKind, &calendar.Event{
			Message: p.Title,

			Start: st,
			End:   et,

			UserID: int(p.UserId),
			PostID: int(p.Id),

			URL:         p.Link,
			Description: render(),
		})
	}
}

// 暂定计划：第 1、3、7、14 天，第 1、3、6 个月，第 1、3、5、10、20、30、40、50 年。
func yieldTimes(t time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		for _, n := range []int{1, 3, 7, 14} {
			t2 := t.AddDate(0, 0, n)
			if !yield(t2) {
				return
			}
		}
		for _, n := range []int{1, 3, 6} {
			t2 := solar.AddMonths(t, n)
			if !yield(t2) {
				return
			}
		}
		for _, n := range []int{1, 3, 5, 10, 20, 30, 40, 50} {
			t2 := solar.AddYears(t, n)
			if !yield(t2) {
				return
			}
		}
	}
}
