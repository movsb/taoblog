package reminders

import (
	"context"
	"fmt"
	"log"
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
}

func NewTask(ctx context.Context, storage utils.PluginStorage, svc proto.TaoBlogServer) *Task {
	t := &Task{
		ctx:   ctx,
		svc:   svc,
		store: storage,
		sched: NewScheduler(),
	}
	go t.run(ctx)
	return t
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
}

func (t *Task) runOnce(ctx context.Context) error {
	now := time.Now()

	ps, err := t.getUpdatedPosts(ctx)
	if err != nil {
		// log.Println(`Reminders.Task.run:`, err)
		return err
	}

	for _, p := range ps {
		log.Println(`处理文章：`, p.Id, p.Title)
		rs, err := t.parsePost(p)
		if err != nil {
			return err
		}
		t.sched.DeleteRemindersByTags(fmt.Sprintf(`post_id:%d`, p.Id))
		for _, r := range rs {
			r.tags = append(r.tags,
				fmt.Sprintf(`post_id:%d`, p.Id),
			)
			if err := t.sched.AddReminder(r, t.notify); err != nil {
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
