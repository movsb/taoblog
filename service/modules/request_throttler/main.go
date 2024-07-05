package request_throttler

import (
	"context"
	"net/netip"
	"time"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	ac := auth.Context(ctx)
	key := throttlerKeyOf(ctx)
	ti, ok := methodThrottlerInfo[info.FullMethod]
	if ok {
		if ti.Interval > 0 {
			if t.throttler.Throttled(key, ti.Interval, false) {
				msg := utils.IIF(ti.Message != "", ti.Message, `你被节流了，请稍候再试。You've been throttled.`)
				return nil, status.Error(codes.Aborted, msg)
			}
		}
		isFromGateway := func() bool {
			md, _ := metadata.FromIncomingContext(ctx)
			if md == nil {
				return false
			}
			ss := md.Get(`X-TaoBlog-Gateway`)
			return len(ss) > 0 && ss[0] == `1`
		}()
		// 可以排除 grpc-client 的调用。
		if ti.Internal && (isFromGateway && !ac.User.IsAdmin()) {
			return nil, status.Error(codes.FailedPrecondition, `此接口限管理员或内部调用。`)
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

	// 是否应该保留为内部调用接口。
	// 限制接口应该尽量被内部调用。
	// 如果不是，也不严重，无权限问题），只是没必要暴露。
	// 主要是对外非管理员接口，管理员接口不受此限制。
	Internal bool
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
	`/protocols.TaoBlog/GetComment`: {
		Internal: true,
	},
	`/protocols.TaoBlog/ListComments`: {
		Internal: true,
	},
	`/protocols.TaoBlog/GetPostComments`: {
		Internal: true,
	},
	`/protocols.TaoBlog/CheckCommentTaskListItems`: {
		Interval:  time.Second * 5,
		Message:   `任务完成得过于频繁？`,
		OnSuccess: false,
	},
	`/protocols.TaoBlog/GetPost`: {
		Internal: true,
	},
	`/protocols.TaoBlog/ListPosts`: {
		Internal: true,
	},
	`/protocols.TaoBlog/GetPostsByTags`: {
		Internal: true,
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
