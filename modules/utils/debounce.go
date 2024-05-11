package utils

import (
	"sync"
	"time"

	"github.com/phuslu/lru"
)

// 防抖神器。
//
// 仅在动作稳定一定时间后才执行。
//
// 比如：我需要在模板文件更新后刷新模板，但是切分支的时候会有收到大量文件修改通知，
// 此时不能每收到一个文件就触发一次刷新，需要在“稳定”（分支切换完全后、几秒内没有文件再有修改）
// 的情况下方可刷新。可避免掉大量无意义的刷新。
//
// NOTE：回调函数 fn 是在独立的线程中被调用的。
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

// 节流神器。
//
// 仅在动作完成一定时间后才允许再度执行。
// 基于时间，而不是漏桶🪣或令牌。
//
// 比如：十秒钟内只允许评论一次。
func NewThrottler[Key comparable]() *Throttler[Key] {
	t := &Throttler[Key]{
		// TODO：不是很清楚这个容量满了会是怎样？
		// 如果满了就被迫删除，那岂不是仍然可以通过刷入大量 key 的
		// 情况下强制失效后使旧 key 再度合法？
		cache: lru.NewTTLCache[Key, time.Time](1024),
	}
	return t
}

type Throttler[Key comparable] struct {
	cache *lru.TTLCache[Key, time.Time]
}

// 检测并更新
func (t *Throttler[Key]) Throttled(key Key, interval time.Duration) bool {
	last, ok := t.cache.Get(key)
	if ok && time.Since(last) < interval {
		return true
	}
	t.cache.Set(key, time.Now(), interval)
	return false
}
