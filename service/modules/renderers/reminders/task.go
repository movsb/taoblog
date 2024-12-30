package reminders

import (
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

type Task struct {
	invalidate func(pid int64)
	posts      map[int64][]*Reminder
	lock       sync.Mutex
}

func NewTask(storage utils.PluginStorage, invalidate func(pid int64)) *Task {
	t := &Task{
		invalidate: invalidate,
		posts:      map[int64][]*Reminder{},
	}
	go t.invalidatePosts()
	return t
}

func (t *Task) AddReminder(pid int64, r *Reminder) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.posts[pid] = append(t.posts[pid], r)
}

func (t *Task) Clear(pid int64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.posts, pid)
}

func (t *Task) invalidatePosts() {
	t.lock.Lock()
	defer t.lock.Unlock()
	for now := range time.Tick(time.Second) {
		// TODO 这个时间可能不精确。比如进程挂起后？
		if now.Hour() == 23 && now.Minute() == 59 && now.Second() == 59 {
			for id := range t.posts {
				t.invalidate(id)
			}
		}
	}
}
