package genealogy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/calendar/solar"
	"github.com/movsb/taoblog/service/modules/renderers"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
)

var calKind = calendar.RegisterKind(func(e *calendar.Event) string {
	name := e.Tags[`name`].(string)
	solar := e.Tags[`solar`].(bool)
	return fmt.Sprintf(`%s\0%s\0%v`, e.Message, name, solar)
})

type RuntimeConfig struct {
	RefreshNow bool `yaml:"refresh_now"`

	refreshNow chan struct{}
	config.Saver
}

func (c *RuntimeConfig) AfterSet(paths config.Segments, obj any) {
	switch paths.At(0).Key {
	case `refresh_now`:
		c.refreshNow <- struct{}{}
	}
}

type Task struct {
	ctx   context.Context
	svc   proto.TaoBlogServer
	store utils.PluginStorage
	cal   *calendar.CalenderService
	rc    *RuntimeConfig

	// 包含族谱图的文章编号
	posts []_Post
	lock  sync.Mutex
}

type _Post struct {
	PostID      int
	UserID      int
	Individuals []*Individual
}

func NewTask(ctx context.Context,
	svc proto.TaoBlogServer,
	store utils.PluginStorage,
	cal *calendar.CalenderService,
) *Task {
	t := &Task{
		ctx:   ctx,
		svc:   svc,
		store: store,
		cal:   cal,
		posts: nil,
		rc: &RuntimeConfig{
			refreshNow: make(chan struct{}),
		},
	}
	if r := runtime_config.FromContext(ctx); r != nil {
		r.Register(`genealogy`, t.rc)
	}
	go t.load()
	go t.run(ctx)
	go t.refreshCalendar(ctx)
	return t
}

func (t *Task) load() {
	str, err := t.store.GetStringDefault(`posts`, `[]`)
	if err != nil {
		log.Println(`Genealogy.Task.load:`, err)
		return
	}
	var posts []int
	if err := json.Unmarshal([]byte(str), &posts); err != nil {
		log.Println(`Genealogy.Task.load:`, err)
		return
	}

	for _, p := range posts {
		pp, err := t.svc.GetPost(
			user.SystemForLocal(t.ctx),
			&proto.GetPostRequest{
				Id: int32(p),
				GetPostOptions: &proto.GetPostOptions{
					WithUserPerms: true,
				},
			},
		)
		if err != nil {
			// 可能是不存在的文章，直接跳过。
			// TODO 需要删除 []posts。
			log.Println(`Genealogy.Task.load:`, err)
			continue
		}
		_, err = t.processSingle(pp, true)
		if err != nil {
			log.Println(`Genealogy.Task.load:`, err)
			continue
		}
	}
	if len(posts) > 0 {
		log.Println(`Genealogy.Task.load:`, `加载完成`)
	}
}

func (t *Task) save(removedPosts []int) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.posts = slices.DeleteFunc(t.posts, func(p _Post) bool {
		return slices.Contains(removedPosts, p.PostID)
	})
	var posts []int
	for _, p := range t.posts {
		posts = append(posts, p.PostID)
	}
	str, _ := json.Marshal(posts)
	return t.store.SetString(`posts`, string(str))
}

func (t *Task) run(ctx context.Context) {
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-time.After(time.Minute):
			if err := t.runOnce(ctx); err != nil {
				log.Println(`族谱图:`, err)
			}
		case <-t.rc.refreshNow:
			if err := t.runOnce(ctx); err != nil {
				log.Println(`族谱图:`, err)
			}
		}
	}
}

func (t *Task) runOnce(ctx context.Context) error {
	ps, err := t.getUpdatedPosts(ctx)
	if err != nil {
		return err
	}
	if len(ps) <= 0 {
		return nil
	}

	var (
		old []int
	)

	for _, p := range ps {
		individuals, err := t.processSingle(p, false)
		if err != nil {
			log.Println(`Genealogy.Task.run:`, err)
			return err
		}
		if len(individuals) <= 0 {
			old = append(old, int(p.Id))
		}
	}

	if err := t.save(old); err != nil {
		log.Println(`Genealogy.Task.run:`, err)
		return err
	}

	// 前面在没有文章的时候提前退出了，此处不需要更新。
	now := time.Now().Unix()
	// log.Println(`当前时间：`, now)
	t.store.SetInteger(lastCheckTimeName, now)

	return nil
}

func (t *Task) processSingle(p *proto.Post, silent bool) (_ []*Individual, outErr error) {
	defer utils.CatchAsError(&outErr)

	individuals := utils.Must1(t.parsePost(p))

	t.cal.Remove(calKind, func(e *calendar.Event) bool {
		return e.PostID == int(p.Id)
	})

	addCache := func(pid, uid int, individuals []*Individual) {
		t.lock.Lock()
		defer t.lock.Unlock()
		t.posts = append(t.posts, _Post{
			PostID:      pid,
			UserID:      uid,
			Individuals: individuals,
		})
	}
	addCache(int(p.Id), int(p.UserId), individuals)
	if p.Status == models.PostStatusPartial {
		for _, up := range p.UserPerms {
			if up.CanRead {
				addCache(int(p.Id), int(up.UserId), individuals)
			}
		}
	}

	for _, individual := range individuals {
		t.addEvent(individual, int(p.Id), int(p.UserId))

		// 如果有分享用户，同时添加到分享用户。
		if p.Status == models.PostStatusPartial {
			for _, up := range p.UserPerms {
				if up.CanRead {
					t.addEvent(individual, int(p.Id), int(up.UserId))
				}
			}
		}

		if !silent {
			log.Println(`族谱：处理完成：`, p.Title, p.Id)
		}
	}

	return individuals, nil
}

// 只添加，不处理旧数据。
func (t *Task) addEvent(ind *Individual, postID, userID int) {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	type Pair struct {
		N     int
		T     time.Time
		Solar bool
	}

	addOnlyValid := func(all ...Pair) {
		times := []Pair{}
		for _, a := range all {
			if !a.T.IsZero() {
				times = append(times, a)
			}
		}
		for _, p := range times {
			start, end := solar.AllDay(p.T)
			t.cal.AddEvent(calKind, &calendar.Event{
				Message: fmt.Sprintf(`%s 的 %d 岁生日🎂️`, ind.Name, p.N),
				Start:   start,
				End:     end,
				PostID:  postID,
				UserID:  userID,
				Tags: map[string]any{
					// 用于去重。
					`name`:  ind.Name,
					`solar`: p.Solar,
				},
			})
		}
	}

	if st := time.Time(ind.Birth.Solar); !st.IsZero() {
		var last, next Pair

		// 可能误写了，去掉。
		st = time.Date(st.Year(), st.Month(), st.Day(), 0, 0, 0, 0, st.Location())

		for i := 0; ; i++ {
			t := solar.AddYears(st, i)
			if t.Before(now) {
				last = Pair{i, t, true}
				continue
			}
			next = Pair{i, t, true}
			break
		}

		addOnlyValid(last, next)
	}

	if !ind.Birth.Lunar.IsZero() {
		var times []Pair

		for i := 0; ; i++ {
			t := ind.Birth.Lunar.AddYears(i)

			// 农历没有具体时间，可忽略去时分秒操作。

			for _, ts := range t {
				times = append(times, Pair{i, ts.SolarTime(), false})
			}

			if !t[0].SolarTime().Before(now) || len(t) > 1 && !t[1].SolarTime().Before(now) {
				break
			}
		}

		// 只保留最多 4 个日子（去年不闰与闰、今年不闰与闰）。
		if n := len(times); n > 4 {
			times = times[n-4:]
		}

		addOnlyValid(times...)
	}
}

const lastCheckTimeName = `last_check_time`

func (t *Task) getUpdatedPosts(ctx context.Context) ([]*proto.Post, error) {
	lastCheckTime, err := t.store.GetIntegerDefault(lastCheckTimeName, 0)
	if err != nil {
		return nil, err
	}

	// log.Println(`上次时间：`, lastCheckTime)

	rsp, err := t.svc.ListPosts(user.SystemForLocal(ctx), &proto.ListPostsRequest{
		ModifiedNotBefore: int32(lastCheckTime),
		GetPostOptions: &proto.GetPostOptions{
			WithUserPerms: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return rsp.Posts, nil
}

func (t *Task) parsePost(p *proto.Post) ([]*Individual, error) {
	if p.SourceType != `markdown` {
		return nil, nil
	}

	var out []*Individual
	md := renderers.NewMarkdown(
		renderers.WithFencedCodeBlockRenderer(`genealogy`, New(
			WithOutput(&out),
			WithoutRender(),
		)),
	)
	_, err := md.Render(p.Source)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (t *Task) refreshCalendar(ctx context.Context) {
	utils.AtMiddleNight(ctx, func() {
		log.Println(`刷新族谱日历：`, time.Now())

		t.cal.Remove(calKind, func(e *calendar.Event) bool {
			return true
		})

		t.lock.Lock()
		defer t.lock.Unlock()

		for _, post := range t.posts {
			for _, ind := range post.Individuals {
				t.addEvent(ind, post.PostID, post.UserID)
			}
		}
	})
}
