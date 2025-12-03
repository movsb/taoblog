package server_auth

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/movsb/taoblog/modules/auth/cookies"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Auth interface {
	AuthRequest(w http.ResponseWriter, r *http.Request) *user.User
	AuthCookie(login, userAgent string) (*user.User, bool)
	GetUserByToken(id int, token string) (*user.User, error)
}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) SetAuth(a Auth) {
	m.a = a
}

type Middleware struct {
	a Auth
}

// 把 Cookie/Authorization 转换成已登录用户。
//
// 适用于浏览器登录的用户。
//
// Note: Cookie/Authorization 同样会被带给 Grpc Gateway，在那里通过 Interceptor 转换成用户。
//
// 纵使本博客程序的 Gateway 和 Service 写在同一个进程，从而允许传递指针。
// 但是这样违背设计原则的使用场景并不被推崇。如果后期有计划拆分成微服务，则会导致改动较多。
func (m *Middleware) UserFromCookieHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := m.a.AuthRequest(w, r)
		remoteAddr := parseRemoteAddrFromHeader(r.Header, r.RemoteAddr)
		userAgent := r.Header.Get(`User-Agent`)
		ac := user.NewContext(r.Context(), u, remoteAddr, userAgent)
		h.ServeHTTP(w, r.WithContext(ac))
	})
}

// 把 Gateway 的 Cookie 转换成已登录用户。
// 适用于服务端代码功能。
//
// NOTE：grpc 服务是 listen 到端口的，和 client 之间只能通过 context 传递的只有 metadata。
// 而 metadata 只是一个普通的 map[string][]string，不能传递指针。
// 纵使本博客程序的 Gateway 和 Service 写在同一个进程，从而允许传递指针。
// 但是这样违背设计原则的使用场景并不被推崇。如果后期有计划拆分成微服务，则会导致改动较多。
func (m *Middleware) UserFromGatewayUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = m.addUserContextToInterceptorForGateway(ctx)
		return handler(ctx, req)
	}
}

func (m *Middleware) UserFromGatewayStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wss := grpc_middleware.WrappedServerStream{
			ServerStream:   ss,
			WrappedContext: m.addUserContextToInterceptorForGateway(ss.Context()),
		}
		return handler(srv, &wss)
	}
}

// TODO 没更改的话不要改变 ServerStream 的 context。
func (m *Middleware) addUserContextToInterceptorForGateway(ctx context.Context) context.Context {
	if user.HasContext(ctx) {
		return ctx
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		panic(status.Error(codes.InvalidArgument, "需要 Metadata。"))
	}

	if len(md.Get(`Authorization`)) > 0 {
		return ctx
	}

	var (
		login     string
		userAgent string
	)

	if cs := md.Get(user.GatewayCookie); len(cs) > 0 {
		header := http.Header{}
		for _, cookie := range cs {
			header.Add(`Cookie`, cookie)
		}
		if loginCookie, err := (&http.Request{Header: header}).Cookie(cookies.CookieNameLogin); err == nil {
			login = loginCookie.Value
		}
	}

	if userAgents := md.Get(user.GatewayUserAgent); len(userAgents) > 0 {
		userAgent = userAgents[0]
	}

	u, _ := m.a.AuthCookie(login, userAgent)

	remoteAddr := parseRemoteAddrFromMetadata(ctx, md)

	return user.NewContext(ctx, u, remoteAddr, userAgent)
}

// 把 Client 的 Token 转换成已登录用户。
// 适用于服务端代码功能。
func (m *Middleware) UserFromClientTokenUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = addUserContextToInterceptorForToken(ctx, func(id int, token string) *user.User {
			u, err := m.a.GetUserByToken(id, token)
			if err == nil {
				return u
			}
			return user.Guest
		})
		return handler(ctx, req)
	}
}

func (m *Middleware) UserFromClientTokenStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wss := grpc_middleware.WrappedServerStream{
			ServerStream: ss,
			WrappedContext: addUserContextToInterceptorForToken(ss.Context(), func(id int, token string) *user.User {
				u, err := m.a.GetUserByToken(id, token)
				if err == nil {
					return u
				}
				return user.Guest
			}),
		}
		return handler(srv, &wss)
	}
}

// TODO 密码错误的时候返回错误而不是游客。
func addUserContextToInterceptorForToken(ctx context.Context, userByToken func(id int, token string) *user.User) context.Context {
	if user.HasContext(ctx) {
		return ctx
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		panic(status.Error(codes.InvalidArgument, "需要 Metadata。"))
	}

	authorizations := (md.Get(`Authorization`))
	if len(authorizations) <= 0 {
		return ctx
	}

	id, token, ok := cookies.ParseAuthorization(authorizations[0])
	if !ok {
		return ctx
	}

	u := userByToken(id, token)

	remoteAddr := parseRemoteAddrFromMetadata(ctx, md)

	var userAgent string
	if userAgents := md.Get(`User-Agent`); len(userAgents) > 0 { // 会自动小写
		userAgent = userAgents[0]
	}

	return user.NewContext(ctx, u, remoteAddr, userAgent)
}

// grpc 服务是被代理过的，所以从 peer.Peer 拿到的是错误的。
// 需要从 nginx 的 forward 中取，得确保配置了。
// NOTE：本地也是统一走 nginx 代理的，也不能缺少。
// TODO：允许本地不走代理，使用 peer.Peer 地址。
// https://en.wikipedia.org/wiki/X-Forwarded-For#Format
// https://github.com/grpc-ecosystem/grpc-gateway/blob/20f268a412e5b342ebfb1a0eef7c3b7bd6c260ea/runtime/context.go#L103
// TODO md 也是从 ctx 来的。
//
// 此 Header 未必可信。
// https://httptoolkit.com/blog/what-is-x-forwarded-for/#can-you-trust-x-forwarded-for
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

// TODO x-forwarded-for 可能是伪造的
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
