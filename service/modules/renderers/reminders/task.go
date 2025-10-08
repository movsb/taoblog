package reminders

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/renderers"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
)

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
	sched *Scheduler
	rc    *RuntimeConfig

	invalidatePost func(id int)

	// 包含日历的文章编号。
	posts map[int]struct{}
	lock  sync.Mutex
}

// TODO invalidate 并不会导致浏览器缓存失效。
func NewTask(ctx context.Context, svc proto.TaoBlogServer,
	invalidatePost func(id int),
	store utils.PluginStorage,
	cal *calendar.CalenderService,
) *Task {
	t := &Task{
		ctx:   ctx,
		svc:   svc,
		store: store,
		sched: NewScheduler(cal, time.Now),
		rc: &RuntimeConfig{
			refreshNow: make(chan struct{}),
		},

		invalidatePost: invalidatePost,
		posts:          make(map[int]struct{}),
	}
	if r := runtime_config.FromContext(ctx); r != nil {
		r.Register(`reminders`, t.rc)
	}
	go t.load()
	go t.run(ctx)
	go t.refreshPosts(ctx)
	go t.refreshCalendar(ctx)
	return t
}

func (t *Task) load() {
	str, err := t.store.GetStringDefault(`posts`, `[]`)
	if err != nil {
		log.Println(`Reminders.Task.load:`, err)
		return
	}
	var posts []int
	if err := json.Unmarshal([]byte(str), &posts); err != nil {
		log.Println(`Reminders.Task.load:`, err)
		return
	}

	t.lock.Lock()
	defer t.lock.Unlock()

	for _, p := range posts {
		t.posts[p] = struct{}{}
	}

	for _, p := range posts {
		pp, err := t.svc.GetPost(
			auth.SystemForLocal(t.ctx),
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
			log.Println(`Reminders.Task.load:`, err)
			continue
		}
		if _, err := t.processSingle(pp, true); err != nil {
			log.Println(`Reminders.Task.load:`, err)
			continue
		}
	}
	if len(posts) > 0 {
		log.Println(`Reminders.Task.load:`, `加载完成`)
	}
}

func (t *Task) save(new, old []int) error {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, n := range new {
		t.posts[n] = struct{}{}
	}
	for _, o := range old {
		delete(t.posts, o)
	}
	var posts []int
	for p := range t.posts {
		posts = append(posts, p)
	}
	str, _ := json.Marshal(posts)
	return t.store.SetString(`posts`, string(str))
}

func (t *Task) run(ctx context.Context) {
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-t.rc.refreshNow:
			if err := t.runOnce(ctx); err != nil {
				log.Println(`提醒:`, err)
			}
		case <-time.After(time.Minute):
			if err := t.runOnce(ctx); err != nil {
				log.Println(`提醒:`, err)
			}
		}
	}
}

func (t *Task) runOnce(ctx context.Context) error {
	ps, err := t.getUpdatedPosts(ctx)
	if err != nil {
		// log.Println(`Reminders.Task.run:`, err)
		return err
	}
	if len(ps) <= 0 {
		return nil
	}

	var (
		new []int
		old []int
	)

	for _, p := range ps {
		found, err := t.processSingle(p, false)
		if err != nil {
			log.Println(`Reminders.Task.run:`, err)
			return err
		}
		if found {
			new = append(new, int(p.Id))
		} else {
			old = append(old, int(p.Id))
		}
	}

	if len(new)+len(old) > 0 {
		if err := t.save(new, old); err != nil {
			log.Println(`Reminders.Task.run:`, err)
			return err
		}
	}

	// 前面在没有文章的时候提前退出了，此处不需要更新。
	now := time.Now().Unix()
	// log.Println(`当前时间：`, now)
	t.store.SetInteger(lastCheckTimeName, now)

	return nil
}

func (t *Task) processSingle(p *proto.Post, silent bool) (_ bool, outErr error) {
	defer utils.CatchAsError(&outErr)
	rs := utils.Must1(t.parsePost(p))
	t.sched.DeleteRemindersByPostID(int(p.Id))
	for _, r := range rs {
		utils.Must(t.sched.AddReminder(int(p.Id), int(p.UserId), r))

		// 如果有分享用户，同时添加到分享用户。
		if p.Status == models.PostStatusPartial {
			for _, up := range p.UserPerms {
				if up.CanRead {
					utils.Must(t.sched.AddReminder(int(p.Id), int(up.UserId), r))
				}
			}
		}

		if !silent {
			log.Println(`提醒：处理完成：`, r.Title, p.Modified)
		}
	}
	return len(rs) > 0, nil
}

const lastCheckTimeName = `last_check_time`

func (t *Task) getUpdatedPosts(ctx context.Context) ([]*proto.Post, error) {
	lastCheckTime, err := t.store.GetIntegerDefault(lastCheckTimeName, 0)
	if err != nil {
		return nil, err
	}

	// log.Println(`上次时间：`, lastCheckTime)

	rsp, err := t.svc.ListPosts(auth.SystemForLocal(ctx), &proto.ListPostsRequest{
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

func (t *Task) parsePost(p *proto.Post) ([]*Reminder, error) {
	if p.SourceType != `markdown` {
		return nil, nil
	}

	var out []*Reminder
	md := renderers.NewMarkdown(
		renderers.WithFencedCodeBlockRenderer(`reminder`, New(WithOutputReminders(&out))),
	)
	_, err := md.Render(p.Source)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// 每天凌晨刷新文章缓存。
func (t *Task) refreshPosts(ctx context.Context) {
	utils.AtMiddleNight(ctx, func() {
		log.Println(`刷新文章提醒缓存：`, time.Now())
		// 等下面的刷新日历先执行。
		time.Sleep(time.Second * 5)
		t.sched.ForEachPost(func(id int) {
			t.invalidatePost(id)
			log.Println(`刷新文章缓存：`, id)
		})
	})
}

func (t *Task) refreshCalendar(ctx context.Context) {
	utils.AtMiddleNight(ctx, func() {
		log.Println(`刷新日历：`, time.Now())
		t.sched.UpdateDaily()
		t.sched.UpdateEvery()
	})
}
