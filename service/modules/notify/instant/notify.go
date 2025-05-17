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

func (n *NotifyLogger) SetPullInterval(d time.Duration) {
	if d <= time.Millisecond*100 {
		d = time.Millisecond * 100
	}
	n.pullInterval = d
}

func (n *NotifyLogger) Notify(deviceKey string, msg Message) error {
	n.store.CreateLog(context.Background(), ty, sty, 1, _Message{
		Message:   msg,
		DeviceKey: deviceKey,
	})
	return nil
}

func (n *NotifyLogger) NotifyImmediately(deviceKey string, msg Message) {
	SendBarkMessage(context.Background(), deviceKey, msg)
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
		if err := SendBarkMessage(ctx, msg.DeviceKey, msg.Message); err != nil {
			log.Println(`NotifyError:`, err)
			time.Sleep(n.pullInterval)
			continue
		} else {
			n.store.DeleteLog(ctx, l.ID)
		}
	}
}
