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

func TestThrottler(t *testing.T) {
	l := utils.NewThrottler[int]()

	expect := func(b, e bool) {
		if b != e {
			t.Fatal(b, e)
		}
	}

	expect(l.Throttled(0, time.Second), false)
	expect(l.Throttled(0, time.Second), true)
	time.Sleep(time.Millisecond * 100)
	expect(l.Throttled(0, time.Second), true)
	time.Sleep(time.Second)
	expect(l.Throttled(0, time.Second), false)
}
