package auth

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type ctxAuthKey struct{}

type AuthContext struct {
	// 当前请求所引用的用户。
	// 不会随不同的请求改变。
	User *User

	// 请求来源 IP 地址。
	// 包括 HTTP 请求，GRPC 请求。
	RemoteAddr netip.Addr

	// 用户使用的代理端名字。
	UserAgent string
}

// 从 Context 里面提取出当前的用户信息。
// 会默认添加 Guest，如果不存在的话。
//
// Note：在当前的实现下，非登录用户/无权限用户被表示为 Guest（id==0）的用户。所以此函数的返回值始终不为空。
// TODO：是不是应该返回 AuthContext 整体？可能包含用户的 IP 地址。
func Context(ctx context.Context) *AuthContext {
	if ac := _Context(ctx); ac != nil {
		return ac
	}
	// Context 可能包含当前请求相关的信息，所以是新建的。
	// TODO：不应该走到这里来，应该 Panic，中间件一定会添加。
	return &AuthContext{User: guest}
}

// 只获取不添加默认。
func _Context(ctx context.Context) *AuthContext {
	if value, ok := ctx.Value(ctxAuthKey{}).(*AuthContext); ok {
		return value
	}
	return nil
}

// 创建一个新的 Context，包含相关信息。
func _NewContext(parent context.Context, user *User, remoteAddr netip.Addr, userAgent string) context.Context {
	ac := AuthContext{
		User:       user,
		RemoteAddr: remoteAddr,
		UserAgent:  userAgent,
	}
	if !remoteAddr.IsValid() {
		panic("无效的远程地址。")
	}
	return context.WithValue(parent, ctxAuthKey{}, &ac)
}

// 把 Cookie 转换成已登录用户。
// 适用于浏览器登录的用户。
//
// Note: Cookie 同样会被带给 Grpc Gateway，在那里通过 Interceptor 转换成用户。
// 纵使本博客程序的 Gateway 和 Service 写在同一个进程，从而允许传递指针。
// 但是这样违背设计原则的使用场景并不被推崇。如果后期有计划拆分成微服务，则会导致改动较多。
func (a *Auth) UserFromCookieHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := a.AuthRequest(r)
		remoteAddr := parseRemoteAddrFromHeader(r.Header, r.RemoteAddr)
		userAgent := r.Header.Get(`User-Agent`)
		ac := _NewContext(r.Context(), user, remoteAddr, userAgent)
		h.ServeHTTP(w, r.WithContext(ac))
	})
}

const (
	GatewayCookie    = runtime.MetadataPrefix + "cookie"
	GatewayUserAgent = runtime.MetadataPrefix + "user-agent"
)

// 把 Gateway 的 Cookie 转换成已登录用户。
// 适用于服务端代码功能。
//
// NOTE：grpc 服务是 listen 到端口的，和 client 之间只能通过 context 传递的只有 metadata。
// 而 metadata 只是一个普通的 map[string][]string，不能传递指针。
// 纵使本博客程序的 Gateway 和 Service 写在同一个进程，从而允许传递指针。
// 但是这样违背设计原则的使用场景并不被推崇。如果后期有计划拆分成微服务，则会导致改动较多。
func (a *Auth) UserFromGatewayUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = a.addUserContextToInterceptorForGateway(ctx)
		return handler(ctx, req)
	}
}

func (a *Auth) UserFromGatewayStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wss := grpc_middleware.WrappedServerStream{
			ServerStream:   ss,
			WrappedContext: a.addUserContextToInterceptorForGateway(ss.Context()),
		}
		return handler(srv, &wss)
	}
}

// TODO 没更改的话不要改变 ServerStream 的 context。
func (a *Auth) addUserContextToInterceptorForGateway(ctx context.Context) context.Context {
	if ac := _Context(ctx); ac != nil {
		return ctx
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		panic(status.Error(codes.InvalidArgument, "需要 Metadata。"))
	}

	// TODO 这是一个手写的常量。
	if len(md.Get(`request_from_gateway`)) <= 0 {
		return ctx
	}

	var (
		login     string
		userAgent string
	)

	if cookies := md.Get(GatewayCookie); len(cookies) > 0 {
		header := http.Header{}
		for _, cookie := range cookies {
			header.Add(`Cookie`, cookie)
		}
		if loginCookie, err := (&http.Request{Header: header}).Cookie(CookieNameLogin); err == nil {
			login = loginCookie.Value
		}
	}

	if userAgents := md.Get(GatewayUserAgent); len(userAgents) > 0 {
		userAgent = userAgents[0]
	}

	user := a.AuthCookie(login, userAgent)

	remoteAddr := parseRemoteAddrFromMetadata(ctx, md)

	return _NewContext(ctx, user, remoteAddr, userAgent)
}

const TokenName = `token`

// 把 Client 的 Token 转换成已登录用户。
// 适用于服务端代码功能。
func (a *Auth) UserFromClientTokenUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = addUserContextToInterceptorForToken(ctx, a.cfg.Key)
		return handler(ctx, req)
	}
}

func (a *Auth) UserFromClientTokenStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wss := grpc_middleware.WrappedServerStream{
			ServerStream:   ss,
			WrappedContext: addUserContextToInterceptorForToken(ss.Context(), a.cfg.Key),
		}
		return handler(srv, &wss)
	}
}

func addUserContextToInterceptorForToken(ctx context.Context, key string) context.Context {
	if ac := _Context(ctx); ac != nil {
		return ctx
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		panic(status.Error(codes.InvalidArgument, "需要 Metadata。"))
	}

	user := guest
	if tokens, ok := md[TokenName]; ok && len(tokens) > 0 {
		if tokens[0] == key {
			user = admin
		}
	}

	remoteAddr := parseRemoteAddrFromMetadata(ctx, md)

	var userAgent string
	if userAgents := md.Get(`User-Agent`); len(userAgents) > 0 { // 会自动小写
		userAgent = userAgents[0]
	}

	return _NewContext(ctx, user, remoteAddr, userAgent)
}

// grpc 服务是被代理过的，所以从 peer.Peer 拿到的是错误的。
// 需要从 nginx 的 forward 中取，得确保配置了。
// NOTE：本地也是统一走 nginx 代理的，也不能缺少。
// TODO：允许本地不走代理，使用 peer.Peer 地址。
// https://en.wikipedia.org/wiki/X-Forwarded-For#Format
// https://github.com/grpc-ecosystem/grpc-gateway/blob/20f268a412e5b342ebfb1a0eef7c3b7bd6c260ea/runtime/context.go#L103
// TODO md 也是从 ctx 来的。
func parseRemoteAddrFromMetadata(ctx context.Context, md metadata.MD) netip.Addr {
	var f string
	if fs, ok := md[`x-forwarded-for`]; ok && len(fs) > 0 {
		f = fs[0]
	}
	if f == "" {
		if peer, ok := peer.FromContext(ctx); ok {
			f, _, _ = net.SplitHostPort(peer.Addr.String())
		}
	}
	return parseRemoteAddr(f)
}

func parseRemoteAddrFromHeader(hdr http.Header, remoteAddr string) netip.Addr {
	var f string
	if fs := hdr.Values(`x-forwarded-for`); len(fs) > 0 {
		f = fs[0]
	}
	if f == "" {
		f, _, _ = net.SplitHostPort(remoteAddr)
	}
	return parseRemoteAddr(f)
}

func parseRemoteAddr(f string) netip.Addr {
	if f == "" {
		panic(`缺少 X-Forwarded-For / RemoteAddr / Peer 字段。`)
	}
	if p := strings.IndexByte(f, ','); p != -1 {
		f = f[:p]
	}
	return netip.MustParseAddr(f)
}
