package utils_test

import (
	"log"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

func TestDebouncer(t *testing.T) {
	last := time.Now()
	time.Sleep(time.Second)
	d := utils.NewDebouncer(time.Second, func() {
		if time.Since(last) < time.Second {
			t.Fatal(`pre-triggered`)
		}
		log.Println(`triggered`)
	})
	time.Sleep(time.Second)
	d.Enter()
	d.Enter()
	time.Sleep(time.Second)
	d.Enter()
	d.Enter()
	time.Sleep(time.Millisecond * 500)
	d.Enter()
	time.Sleep(time.Millisecond * 500)
	d.Enter()
	time.Sleep(time.Millisecond * 500)
	d.Enter()
	time.Sleep(time.Millisecond * 500)
	d.Enter()
	time.Sleep(time.Second)
}
