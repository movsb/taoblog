package live_check

import (
	"context"
	"expvar"
	"log"
	"runtime"
	"time"

	"github.com/movsb/taoblog/cmd/server/maintenance"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/micros/auth/user"
)

var (
	stack = expvar.NewString(`live_check_stack`)
	last  time.Time
)

// 服务可用性检测。
//
// 函数不会返回，除非 ctx 结束。
//
// NOTE: 文章 1 必须存在。可以是非公开状态。
func LiveCheck(ctx context.Context, svc *service.Service, maintenanceMode maintenance.MaintenanceMode, sendNotify func(title, message string)) {
	// 如果接口可用，返回 true。
	check := func() bool {
		now := time.Now()
		svc.GetPost(user.SystemForLocal(ctx), &proto.GetPostRequest{Id: 1})
		if elapsed := time.Since(now); elapsed > time.Second*10 {
			maintenanceMode.Enter(`我也不知道为什么，反正就是服务接口卡住了🥵。`, -1)
			sendNotify(`服务不可用`, `保活检测卡住了`)
			log.Println(`服务接口响应非常慢了。`)

			// 正式环境时打印完整的栈信息。
			if !version.DevMode() {
				buf := make([]byte, 1<<20)
				runtime.Stack(buf, true)
				stack.Set(string(buf))
				last = time.Now()
			}

			return false
		}

		maintenanceMode.Leave()

		// 最多保留一天的栈。
		if !last.IsZero() && time.Since(last) > time.Hour*24 {
			stack.Set(``)
		}

		return true
	}

	for {
		if !check() {
			time.Sleep(time.Second * 5)
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute * 1):
		}
	}
}
