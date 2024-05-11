package utils

import (
	"sync"
	"time"
)

// 防抖神器。
//
// 仅在动作稳定一定时间后才执行。
//
// 比如：我需要在模板文件更新后刷新模板，但是切分支的时候会有收到大量文件修改通知，
// 此时不能每收到一个文件就触发一次刷新，需要在“稳定”（分支切换完全后、几秒内没有文件再有修改）
// 的情况下方可刷新。可避免掉大量无意义的刷新。
func NewDebouncer(interval time.Duration, fn func()) *_Debouncer {
	if interval < time.Second {
		panic(`invalid debouncer interval`)
	}
	d := &_Debouncer{
		f:          fn,
		remain:     interval,
		interval:   interval,
		resolution: interval / 10,
	}
	return d
}

type _Debouncer struct {
	l          sync.Mutex
	t          *time.Ticker
	f          func()
	remain     time.Duration
	interval   time.Duration
	resolution time.Duration
}

func (d *_Debouncer) Enter() {
	d.l.Lock()
	d.remain = d.interval
	// log.Println("重置倒计时")
	if d.t == nil {
		d.t = time.NewTicker(d.resolution)
		// log.Println("启动倒计时")
		go d.wait()
	}
	d.l.Unlock()
}

func (d *_Debouncer) wait() {
	d.l.Lock()
	c := d.t.C
	d.l.Unlock()
	for {
		func() {
			<-c
			d.l.Lock()
			defer d.l.Unlock()
			d.remain -= d.resolution
			// log.Println("剩余-1")
			if d.remain <= 0 {
				// log.Println("触发")
				d.t.Stop()
				d.t = nil
				d.f()
			}
		}()
	}
}
