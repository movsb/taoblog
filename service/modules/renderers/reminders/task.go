package reminders

import (
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

type Task struct {
	invalidate func(pid int64)
	posts      map[int64]struct{}
}

func NewTask(storage utils.PluginStorage, invalidate func(pid int64)) *Task {
	t := &Task{
		invalidate: invalidate,
	}
	go t.invalidatePosts()
	return t
}

func (t *Task) AddReminder(pid int64) {
	t.posts[pid] = struct{}{}
}

func (t *Task) invalidatePosts() {
	for now := range time.Tick(time.Second) {
		// TODO 这个时间可能不精确。比如进程挂起后？
		if now.Hour() == 23 && now.Minute() == 59 && now.Second() == 59 {
			for id := range t.posts {
				t.invalidate(id)
			}
		}
	}
}
