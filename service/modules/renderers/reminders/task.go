package reminders

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taorm"
)

type Task struct {
	ctx   context.Context
	svc   proto.TaoBlogServer
	store utils.PluginStorage
	sched *Scheduler

	invalidatePost func(id int)
	sendNotify     func(message string)
}

func NewTask(ctx context.Context, svc proto.TaoBlogServer,
	invalidatePost func(id int),
	sendNotify func(message string),
) *Task {
	t := &Task{
		ctx:   ctx,
		svc:   svc,
		store: utils.NewInMemoryStorage(),
		sched: NewScheduler(),

		invalidatePost: invalidatePost,
		sendNotify:     sendNotify,
	}
	go t.run(ctx)
	go t.refreshPosts(ctx)
	return t
}

func (t *Task) CalenderService() http.Handler {
	info, _ := t.svc.GetInfo(t.ctx, &proto.GetInfoRequest{})
	return NewCalendarService(info.Name, t.sched)
}

func (t *Task) run(ctx context.Context) {
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-time.After(time.Second * 10):
			if err := t.runOnce(ctx); err != nil {
				log.Println(`提醒:`, err)
			}
		}
	}
}

func (t *Task) notify(now time.Time, message string) {
	log.Println(`提醒：`, message)
	t.sendNotify(message)
}

func (t *Task) runOnce(ctx context.Context) error {
	now := time.Now()

	ps, err := t.getUpdatedPosts(ctx)
	if err != nil {
		// log.Println(`Reminders.Task.run:`, err)
		return err
	}

	for _, p := range ps {
		// log.Println(`处理文章：`, p.Id, p.Title)
		rs, err := t.parsePost(p)
		if err != nil {
			return err
		}
		t.sched.DeleteRemindersByPostID(int(p.Id))
		for _, r := range rs {
			if err := t.sched.AddReminder(int(p.Id), r, t.notify); err != nil {
				return err
			}
		}
	}

	t.store.Set(lastCheckTimeName, fmt.Sprint(now.Unix()))

	return nil
}

const lastCheckTimeName = `last_check_time`

func (t *Task) getUpdatedPosts(ctx context.Context) ([]*proto.Post, error) {
	lastCheckTimeString, err := t.store.Get(lastCheckTimeName)
	if err != nil {
		if !taorm.IsNotFoundError(err) {
			return nil, err
		}
		lastCheckTimeString = `0`
	}
	lastCheckTime, err := strconv.Atoi(lastCheckTimeString)
	if err != nil {
		return nil, err
	}

	// now := time.Now().Unix()

	rsp, err := t.svc.ListPosts(auth.SystemAdmin(ctx), &proto.ListPostsRequest{
		ContentOptions:    &proto.PostContentOptions{},
		ModifiedNotBefore: int32(lastCheckTime),
	})
	if err != nil {
		return nil, err
	}

	// t.store.Set(lastCheckTimeName, fmt.Sprint(now))

	return rsp.Posts, nil
}

func (t *Task) parsePost(p *proto.Post) ([]*Reminder, error) {
	if p.SourceType != `markdown` {
		return nil, nil
		// return nil, fmt.Errorf(`不支持非 Markdown 类型的文章`)
	}

	var out []*Reminder
	md := renderers.NewMarkdown(
		New(WithOutputReminders(&out)),
	)
	_, err := md.Render(p.Source)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// 每天凌晨刷新文章缓存。
func (t *Task) refreshPosts(ctx context.Context) {
	execute := func() {
		log.Println(`刷新文章提醒缓存：`, time.Now())
		t.sched.ForEachPost(func(id int, jobs []Job) {
			t.invalidatePost(id)
			log.Println(`刷新文章缓存：`, id)
		})
	}

	ticker := time.NewTicker(time.Second * 50)
	defer ticker.Stop()
	last := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if (now.Hour() == 0 && now.Minute() == 0) || (last.Day() != now.Day()) {
				execute()
			}
			last = now
		}
	}
}
