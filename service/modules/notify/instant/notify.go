package instant

import (
	"context"
	"log"
	"time"

	"github.com/movsb/taoblog/modules/logs"
)

type NotifyLogger struct {
	store logs.Logger

	pullInterval time.Duration
}

func NewNotifyLogger(store logs.Logger) *NotifyLogger {
	n := &NotifyLogger{
		store: store,
	}
	n.SetPullInterval(time.Second * 5)
	go n.process(context.Background())
	return n
}

const (
	ty  = `notify`
	sty = `message`
)

type _Message struct {
	Title, Message string

	ChanifyToken string
}

func (n *NotifyLogger) SetPullInterval(d time.Duration) {
	if d <= time.Millisecond*100 {
		d = time.Millisecond * 100
	}
	n.pullInterval = d
}

func (n *NotifyLogger) Notify(title, message string, chanifyToken string) error {
	n.store.CreateLog(context.Background(), ty, sty, 1, _Message{
		Title:   title,
		Message: message,

		ChanifyToken: chanifyToken,
	})
	return nil
}

func (n *NotifyLogger) NotifyImmediately(title, message string, chanifyToken string) {
	NewChanifyNotify(chanifyToken).Notify(title, message)
}

func (n *NotifyLogger) process(ctx context.Context) {
	for {
		var msg _Message
		l := n.store.FindLog(ctx, ty, sty, &msg)
		if l == nil {
			time.Sleep(n.pullInterval)
			continue
		}
		log.Println(`找到日志：`, l.ID)
		ch := NewChanifyNotify(msg.ChanifyToken)
		if err := ch.Notify(msg.Title, msg.Message); err != nil {
			log.Println(`NotifyError:`, err)
			time.Sleep(n.pullInterval)
			continue
		} else {
			n.store.DeleteLog(ctx, l.ID)
		}
	}
}
