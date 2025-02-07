package reminders

import (
	"github.com/movsb/taoblog/modules/utils"
)

type Task struct {
	invalidate func(pid int64)
	posts      map[int64][]*Reminder
}

func NewTask(storage utils.PluginStorage, invalidate func(pid int64)) *Task {
	t := &Task{
		invalidate: invalidate,
		posts:      map[int64][]*Reminder{},
	}
	return t
}
