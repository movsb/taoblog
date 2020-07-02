package auth

// https://github.com/movsb/tiger-backend/blob/master/externals/helpers/cookie.go

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GatewayAuthInterceptor ...
func GatewayAuthInterceptor(a *Auth) grpc.UnaryServerInterceptor {
	const (
		gatewayCookie    = runtime.MetadataPrefix + "cookie"
		gatewayUserAgent = runtime.MetadataPrefix + "user-agent"
	)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			var (
				login     string
				userAgent string
			)
			if cookies := md.Get(gatewayCookie); len(cookies) > 0 {
				header := http.Header{}
				for _, cookie := range cookies {
					header.Add(`Cookie`, cookie)
				}
				if loginCookie, err := (&http.Request{Header: header}).Cookie(`login`); err == nil {
					login = loginCookie.Value
				}
			}
			if userAgents := md.Get(gatewayUserAgent); len(userAgents) > 0 {
				userAgent = userAgents[0]
			}
			ctx = a.AuthCookie3(login, userAgent).Context(ctx)
		}

		return handler(ctx, req)
	}
}
