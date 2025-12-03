package user

import (
	"context"
	"net/netip"

	"github.com/movsb/taoblog/modules/geo/geoip"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxAuthKey struct{}

// 鉴权后保存的用户信息。
// 进程内使用（不含 Gateway）。
type AuthContext struct {
	// 当前请求所引用的用户。
	//
	// 始终不为空；如果是未登录用户，则为 guest。
	User *User

	// 请求来源 IP 地址。
	// 包括 HTTP 请求，GRPC 请求。
	// 始终不为空。
	RemoteAddr netip.Addr

	// RemoteAddr 是否在中国。
	// 用于加速资源访问。
	InChina bool

	// 用户使用的代理端名字。
	UserAgent string
}

// 只获取不添加默认。
func _Context(ctx context.Context) *AuthContext {
	if value, ok := ctx.Value(ctxAuthKey{}).(*AuthContext); ok {
		return value
	}
	return nil
}

// 创建一个新的 Context，包含相关信息。
func NewContext(parent context.Context, user *User, remoteAddr netip.Addr, userAgent string) context.Context {
	ac := AuthContext{
		User:       user,
		RemoteAddr: remoteAddr,
		UserAgent:  userAgent,
	}
	if !remoteAddr.IsValid() {
		panic("无效的远程地址。")
	}
	ac.InChina = geoip.IsInChina(remoteAddr)
	return context.WithValue(parent, ctxAuthKey{}, &ac)
}

// 判断 ctx 中是否已包含 AuthContext。
func HasContext(ctx context.Context) bool {
	return _Context(ctx) != nil
}

// 从 Context 里面提取出当前的用户信息。
//
// Note：在当前的实现下，非登录用户/无权限用户被表示为 Guest（id==0）的用户，
// 所以此函数的返回值始终不为空。因此，如果取不到用户信息，会 panic。
func Context(ctx context.Context) *AuthContext {
	if ac := _Context(ctx); ac != nil {
		return ac
	}
	panic(`Context 中未包含登录用户信息。`)
}

////////////////////////////////////////////////////////////////////////////////

var Localhost = netip.AddrFrom4([4]byte{127, 0, 0, 1})

// 系统管理员身份。相当于后台任务执行者。拥有所有权限。
// 不用 == Admin：一个是真人，一个是拟人。
// 权限可以一样，也可以不一样。
// 比如 System 不允许真实登录，只是后台操作。
// 只能进程内/本地使用，不能跨网络使用（包括 gateway 也不行）。
func SystemForLocal(ctx context.Context) context.Context {
	return NewContext(ctx, System, Localhost, `system_admin`)
}

// 访客身份。
// 只能进程内/本地使用，不能跨网络使用（包括 gateway 也不行）。
func GuestForLocal(ctx context.Context) context.Context {
	return NewContext(ctx, Guest, Localhost, `guest_context`)
}

// 只能用于 Gateway，充当 System 用户。
func SystemForGateway(ctx context.Context) context.Context {
	md := metadata.Pairs(`Authorization`, System.AuthorizationValue())
	return metadata.NewOutgoingContext(ctx, md)
}

const noPerm = `此操作无权限。`

func MustNotBeGuest(ctx context.Context) *AuthContext {
	ac := Context(ctx)
	if !ac.User.IsGuest() {
		return ac
	}
	panic(status.Error(codes.PermissionDenied, noPerm))
}

func MustBeAdmin(ctx context.Context) *AuthContext {
	ac := Context(ctx)
	if ac.User.IsAdmin() || ac.User.IsSystem() {
		return ac
	}
	panic(status.Error(codes.PermissionDenied, noPerm))
}
