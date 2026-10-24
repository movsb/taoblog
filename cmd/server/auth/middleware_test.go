package server_auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/movsb/taoblog/service/micros/auth/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestConflictingCredentials(t *testing.T) {
	for _, login := range []string{"invalid", ""} {
		for _, authorization := range []string{"token 2:invalid", "invalid", ""} {
			header := http.Header{}
			header.Add("Cookie", "taoblog.nickname=test")
			header.Add("Cookie", "taoblog.login="+login)
			header.Add("Authorization", authorization)

			t.Run("HTTP/"+login+"/"+authorization, func(t *testing.T) {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header = header
				w := httptest.NewRecorder()
				// No auth backend is installed: rejection must happen before authentication.
				NewMiddleware().UserFromCookieHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
					t.Fatal("conflicting credentials reached handler")
				})).ServeHTTP(w, r)
				if w.Code != http.StatusBadRequest {
					t.Fatalf("got HTTP %d, want 400", w.Code)
				}
			})

			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
				user.GatewayCookie, "taoblog.nickname=test",
				user.GatewayCookie, "taoblog.login="+login,
				"authorization", authorization,
			))
			t.Run("unary/"+login+"/"+authorization, func(t *testing.T) {
				_, err := NewMiddleware().UserFromGatewayUnaryInterceptor()(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) {
					t.Fatal("conflicting credentials reached RPC handler")
					return nil, nil
				})
				if status.Code(err) != codes.InvalidArgument {
					t.Fatalf("got %v, want InvalidArgument", err)
				}
			})
			t.Run("stream/"+login+"/"+authorization, func(t *testing.T) {
				err := NewMiddleware().UserFromGatewayStreamInterceptor()(nil, &testServerStream{ctx: ctx}, &grpc.StreamServerInfo{}, func(any, grpc.ServerStream) error {
					t.Fatal("conflicting credentials reached stream handler")
					return nil
				})
				if status.Code(err) != codes.InvalidArgument {
					t.Fatalf("got %v, want InvalidArgument", err)
				}
			})
		}
	}
}

func TestNonConflictingCredentials(t *testing.T) {
	for _, header := range []http.Header{
		{},
		{"Cookie": {"taoblog.login=value"}},
		{"Authorization": {"token 2:value"}},
		{"Cookie": {"taoblog.user_id=2; taoblog.nickname=test"}, "Authorization": {"token 2:value"}},
	} {
		if hasConflictingCredentials(header) {
			t.Fatalf("unexpected conflict for %v", header)
		}
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			user.GatewayCookie: header.Values("Cookie"),
			"authorization":    header.Values("Authorization"),
		})
		if gatewayCredentialsConflict(ctx) {
			t.Fatalf("unexpected metadata conflict for %v", header)
		}
	}
}

type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *testServerStream) Context() context.Context { return s.ctx }
