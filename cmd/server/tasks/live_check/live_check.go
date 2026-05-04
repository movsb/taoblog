package live_check

import (
	"context"
	"expvar"
	"log"
	"net/http"
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
//   - 函数不会返回，除非 ctx 结束。
//   - 注意检测的时候都不应该增加首页/文章的阅读次数。
//   - 文章 1 必须存在。可以是非公开状态。
//
// 出现过 Home 因磁盘满后的 502 错误，但是找不到栈，因此也检测主页。
func LiveCheck(ctx context.Context, svc *service.Service, maintenanceMode maintenance.MaintenanceMode, sendNotify func(title, message string)) {
	l := _LiveCheck{
		ctx:         ctx,
		svc:         svc,
		maintenance: maintenanceMode,
		sendNotify:  sendNotify,
	}
	for {
		if !l.checkPost() || !l.checkHome() {
			time.Sleep(time.Second * 5)
			continue
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute * 5):
		}
	}
}

type _LiveCheck struct {
	ctx         context.Context
	svc         *service.Service
	maintenance maintenance.MaintenanceMode
	sendNotify  func(title, message string)
}

// 如果接口可用，返回 true。
func (l *_LiveCheck) checkPost() bool {
	now := time.Now()
	_, err := l.svc.GetPost(user.SystemForLocal(l.ctx), &proto.GetPostRequest{Id: 1})
	if elapsed := time.Since(now); elapsed > time.Second*10 {
		l.maintenance.Enter(`我也不知道为什么，反正就是服务接口卡住了🥵。`, -1)
		l.sendNotify(`服务不可用`, `保活检测卡住了`)
		log.Println(`服务接口响应非常慢了。`)

		// 正式环境时打印完整的栈信息。
		if !version.DevMode() {
			buf := make([]byte, 1<<20)
			for {
				n := runtime.Stack(buf, true)
				if n < len(buf) {
					buf = buf[:n]
					break
				}
				buf = make([]byte, 2*len(buf))
			}
			stack.Set(string(buf))
			last = time.Now()
		}

		return false
	}
	if err != nil {
		l.sendNotify(`获取文章时出错。`, `在线状态检测`)
		return false
	}

	l.maintenance.Leave()

	// 最多保留一天的栈。
	if !last.IsZero() && time.Since(last) > time.Hour*24 {
		stack.Set(``)
	}

	return true
}

func (l *_LiveCheck) checkHome() bool {
	ctx, cancel := context.WithTimeout(l.ctx, time.Second*5)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l.svc.Config().Site.GetHome(), nil)
	if err != nil {
		log.Println(err)
		return false
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		l.sendNotify(`获取首页数据失败：`+err.Error(), `在线状态检测`)
		return false
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		l.sendNotify(`获取首页数据状态码不正确：`+rsp.Status, `在线状态检测`)
		return false
	}
	return true
}
