package instant

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/movsb/taoblog/modules/logs"
)

type NotifyLogger struct {
	store logs.Logger

	notifier     atomic.Value
	pullInterval time.Duration
}

func NewNotifyLogger(store logs.Logger) *NotifyLogger {
	n := &NotifyLogger{
		store: store,
	}
	n.SetPullInterval(time.Second * 5)
	return n
}

const (
	ty  = `notify`
	sty = `message`
)

type _Message struct {
	Title, Message string
}

func (n *NotifyLogger) SetPullInterval(d time.Duration) {
	if d <= time.Millisecond*100 {
		d = time.Millisecond * 100
	}
	n.pullInterval = d
}

func (n *NotifyLogger) SetNotifier(backend Notifier) {
	if old := n.notifier.Swap(backend); old == nil {
		go n.process(context.Background())
	}
}

func (n *NotifyLogger) Notify(title, message string) error {
	n.store.CreateLog(context.Background(), ty, sty, 1, _Message{
		Title:   title,
		Message: message,
	})
	return nil
}

func (n *NotifyLogger) NotifyImmediately(title, message string) {
	backend, _ := n.notifier.Load().(Notifier)
	if backend != nil {
		backend.Notify(title, message)
	}
}

func (n *NotifyLogger) process(ctx context.Context) {
	for {
		var msg _Message
		l := n.store.FindLog(ctx, ty, sty, &msg)
		backend, _ := n.notifier.Load().(Notifier)
		if l != nil && backend != nil {
			log.Println(`找到日志：`, l.ID)
			if err := backend.Notify(msg.Title, msg.Message); err != nil {
				log.Println(`NotifyError:`, err)
				time.Sleep(n.pullInterval)
				continue
			} else {
				n.store.DeleteLog(ctx, l.ID)
			}
		} else {
			time.Sleep(n.pullInterval)
		}
	}
}
