package auth

import (
	"context"
	"net/http"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxAuthKey struct{}

type AuthContext struct {
	User *User
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
	return &AuthContext{User: guest}
}

// 只获取不添加默认。
func _Context(ctx context.Context) *AuthContext {
	if value, ok := ctx.Value(ctxAuthKey{}).(*AuthContext); ok {
		return value
	}
	return nil
}

// 把 Cookie 转换成已登录用户。
// 适用于浏览器登录的用户。
//
// Note: Cookie 同样会被带给 Grpc Gateway，在那里通过 Interceptor 转换成用户。
func (a *Auth) UserFromCookieHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := a.AuthRequest(r)
		ctxUser := user.Context(r.Context())
		ctxReq := r.WithContext(ctxUser)
		h.ServeHTTP(w, ctxReq)
	})
}

// 把 Gateway 的 Cookie 转换成已登录用户。
// 适用于服务端代码功能。
//
// NOTE：grpc 服务是 listen 到端口的，和 client 之间只能通过 context 传递的只有 metadata。
// 而 metadata 只是一个普通的 map[string][]string，不能传递指针。
// 纵使本博客程序的 Gateway 和 Service 写在同一个进程，从而允许传递指针。
// 但是这样违背设计原则的使用场景并不被推崇。如果后期有计划拆分成微服务，则会导致改动较多。
func (a *Auth) UserFromGatewayCookieInterceptor() grpc.UnaryServerInterceptor {
	const (
		gatewayCookie    = runtime.MetadataPrefix + "cookie"
		gatewayUserAgent = runtime.MetadataPrefix + "user-agent"
	)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if ac := _Context(ctx); ac != nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			panic(status.Error(codes.InvalidArgument, "需要 Metadata。"))
		}

		var (
			login     string
			userAgent string
		)

		if cookies := md.Get(gatewayCookie); len(cookies) > 0 {
			header := http.Header{}
			for _, cookie := range cookies {
				header.Add(`Cookie`, cookie)
			}
			if loginCookie, err := (&http.Request{Header: header}).Cookie(CookieNameLogin); err == nil {
				login = loginCookie.Value
			}
		}

		if userAgents := md.Get(gatewayUserAgent); len(userAgents) > 0 {
			userAgent = userAgents[0]
		}

		ctx = a.AuthCookie(login, userAgent).Context(ctx)

		return handler(ctx, req)
	}
}

const TokenName = `token`

// 把 Client 的 Token 转换成已登录用户。
// 适用于服务端代码功能。
func (a *Auth) UserFromClientTokenUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = addUserContextToInterceptor(ctx, a.cfg.Key)
		return handler(ctx, req)
	}
}

func (a *Auth) UserFromClientTokenStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wss := grpc_middleware.WrappedServerStream{
			ServerStream:   ss,
			WrappedContext: addUserContextToInterceptor(ss.Context(), a.cfg.Key),
		}
		return handler(srv, &wss)
	}
}

func addUserContextToInterceptor(ctx context.Context, key string) context.Context {
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

	return user.Context(ctx)
}
