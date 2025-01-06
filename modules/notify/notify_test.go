package notify_test

import (
	"io"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/notify"
)

type _RandomErrorNotify struct{}

func (n *_RandomErrorNotify) Notify(title, message string) error {
	if rand.Int()&1 == 0 {
		return io.EOF
	}
	return nil
}

func TestNotify(t *testing.T) {
	db := server.InitDatabase(``)
	defer db.Close()

	n := notify.NewNotifyLogger(notify.NewLogStore(db))
	n.SetNotifier(&_RandomErrorNotify{})
	n.SetPullInterval(10 * time.Millisecond)
	n.Notify(`title`, `message 1`)
	n.Notify(`title`, `message 2`)
	n.Notify(`title`, `message`)
	n.Notify(`title`, `message`)
	n.Notify(`title`, `message`)
	log.Println(`退出`)
	time.Sleep(time.Second * 2)
}
