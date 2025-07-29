package utils_test

import (
	"log"
	"sync/atomic"
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
	t.SkipNow()

	l := utils.NewThrottler[int]()

	expect := func(n int, b, e bool) {
		if b != e {
			t.Fatal(n, b, e)
		}
	}

	expect(1, l.Throttled(0, time.Second, true), false)
	expect(2, l.Throttled(0, time.Second, true), true)
	time.Sleep(time.Millisecond * 500)
	expect(3, l.Throttled(0, time.Second, true), true)
	time.Sleep(time.Second * 2)
	expect(4, l.Throttled(0, time.Second, true), false)
}

func TestLimitNumberOfProcesses(t *testing.T) {
	var n atomic.Int32
	ch := make(chan struct{}, 5)
	for i := range 5 {
		go utils.LimitNumberOfProcesses(`test`, &n, 3, func() {
			t.Log(`sleeping`)
			time.Sleep(time.Second * 1)
			t.Log(i)
			ch <- struct{}{}
		})
	}
	for range 5 {
		<-ch
	}
}
