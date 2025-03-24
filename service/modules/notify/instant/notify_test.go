package instant_test

import (
	"io"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/service/modules/notify/instant"
	"github.com/movsb/taoblog/setup/migration"
)

type _RandomErrorNotify struct{}

func (n *_RandomErrorNotify) Notify(title, message string) error {
	if rand.Int()&1 == 0 {
		return io.EOF
	}
	return nil
}

func TestNotify(t *testing.T) {
	db := migration.InitPosts(``, false)
	defer db.Close()

	n := instant.NewNotifyLogger(logs.NewLogStore(db))
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
