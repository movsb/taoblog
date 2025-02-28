package theme

import (
	"context"
	"log"
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

// 每次访问一篇文章均需要增加文章访问次数，这样会高频地更新数据库。
// 此这个结构体用于给访问量统计功能“防抖”。
type _IncViewDebouncer struct {
	d *utils.Debouncer

	// map[文章编号]当前增加次数
	m map[int]int

	ch    chan int
	flush chan struct{}
	save  func(m map[int]int)
}

func NewIncViewDebouncer(ctx context.Context, save func(m map[int]int)) *_IncViewDebouncer {
	d := &_IncViewDebouncer{
		m:     map[int]int{},
		ch:    make(chan int, 10),
		flush: make(chan struct{}, 1),
		save:  save,
	}
	d.d = utils.NewDebouncer(time.Second*10, func() {
		select {
		case d.flush <- struct{}{}:
		default:
		}
	})
	go d.run(ctx)
	return d
}

func (d *_IncViewDebouncer) run(ctx context.Context) {
	const forceInterval = time.Minute
	ticker := time.NewTicker(forceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println(`结束`)
			return
		case p := <-d.ch:
			d.m[p]++
		case <-ticker.C:
			select {
			case d.flush <- struct{}{}:
			default:
			}
		case <-d.flush:
			log.Println(`保存文章访问次数`)
			d.save(d.m)
		}
	}
}

func (d *_IncViewDebouncer) Increase(p int) {
	select {
	case d.ch <- p:
		d.d.Enter()
	default:
	}
}
