package request_throttler

import (
	"context"
	"net/netip"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func New() grpc.UnaryServerInterceptor {
	return (&_Throttler{
		throttler: utils.NewThrottler[_RequestThrottlerKey](),
	}).throttlerGatewayInterceptor
}

type _Throttler struct {
	// 请求节流器。
	throttler *utils.Throttler[_RequestThrottlerKey]
}

func (t *_Throttler) throttlerGatewayInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// ac := auth.Context(ctx)
	key := throttlerKeyOf(ctx)
	ti, ok := methodThrottlerInfo[info.FullMethod]
	if ok {
		if ti.Interval > 0 {
			if t.throttler.Throttled(key, ti.Interval, false) {
				msg := utils.IIF(ti.Message != "", ti.Message, `你被节流了，请稍候再试。You've been throttled.`)
				return nil, status.Error(codes.Aborted, msg)
			}
		}
	}

	resp, err := handler(ctx, req)

	if !ti.OnSuccess || err == nil {
		t.throttler.Update(key, ti.Interval)
	}

	return resp, err
}

var methodThrottlerInfo = map[string]struct {
	Interval time.Duration
	Message  string

	// 仅节流返回正确错误码的接口。
	// 如果接口返回错误，不更新。
	OnSuccess bool
}{
	`/protocols.TaoBlog/CreateComment`: {
		Interval:  time.Second * 10,
		Message:   `评论发表过于频繁，请稍等几秒后再试。`,
		OnSuccess: true,
	},
	`/protocols.TaoBlog/UpdateComment`: {
		Interval:  time.Second * 5,
		Message:   `评论更新过于频繁，请稍等几秒后再试。`,
		OnSuccess: true,
	},
	`/protocols.TaoBlog/CheckCommentTaskListItems`: {
		Interval:  time.Second * 5,
		Message:   `任务完成得过于频繁？`,
		OnSuccess: false,
	},
}

// 请求节流器限流信息。
// 由于没有用户系统，目前根据 IP 限流。
// 这样会对网吧、办公网络非常不友好。
type _RequestThrottlerKey struct {
	UserID int
	IP     netip.Addr
	Method string // 指 RPC 方法，用路径代替。
}

func throttlerKeyOf(ctx context.Context) _RequestThrottlerKey {
	ac := auth.Context(ctx)
	method, ok := grpc.Method(ctx)
	if !ok {
		panic(status.Error(codes.Internal, "没有找到调用方法。"))
	}
	return _RequestThrottlerKey{
		UserID: int(ac.User.ID),
		IP:     ac.RemoteAddr,
		Method: method,
	}
}
