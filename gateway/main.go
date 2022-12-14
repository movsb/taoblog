package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/modules/pingback"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

type Gateway struct {
	mux     *http.ServeMux
	service *service.Service
	auther  *auth.Auth
}

func NewGateway(service *service.Service, auther *auth.Auth, mux *http.ServeMux) *Gateway {
	g := &Gateway{
		mux:     mux,
		service: service,
		auther:  auther,
	}

	mux2 := runtime.NewServeMux(
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
			},
		),
	)

	mux.HandleFunc(`/v1/`, deprecated)
	mux.HandleFunc(`/v2/`, deprecated)
	mux.Handle(`/v3/`, mux2)

	if err := g.runHTTPService(context.TODO(), mux, mux2); err != nil {
		panic(err)
	}

	return g
}

// runHTTPService ...
// TODO auth
func (g *Gateway) runHTTPService(ctx context.Context, mux *http.ServeMux, mux2 *runtime.ServeMux) error {
	protocols.RegisterTaoBlogHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), []grpc.DialOption{grpc.WithInsecure()})
	protocols.RegisterSearchHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), []grpc.DialOption{grpc.WithInsecure()})

	compile := func(rule string) httprule.Template {
		if compiler, err := httprule.Parse(rule); err != nil {
			panic(err)
		} else {
			return compiler.Compile()
		}
	}

	handle := func(method string, rule string, handler runtime.HandlerFunc) {
		t := compile(rule)
		pattern, err := runtime.NewPattern(1, t.OpCodes, t.Pool, t.Verb)
		if err != nil {
			panic(err)
		}

		mux2.Handle(method, pattern, handler)
	}

	handle("GET", `/v3/api`, getAPI)
	handle("GET", `/v3/api/swagger`, getSwagger)

	handle(`GET`, `/v3/comments/{id}/avatar`, g.GetAvatar)

	handle(`GET`, `/v3/posts/{post_id}/files`, g.ListFiles)
	handle(`GET`, `/v3/posts/{post_id}/files/{file=**}`, g.GetFile)
	handle(`POST`, `/v3/posts/{post_id}/files/{file=**}`, g.CreateFile)
	// handle(`DELETE`, `/v3/posts/{post_id}/files/{file=**}`, g.DeleteFile)

	pingbackHandler := pingback.Handler(g.service.Pingback)
	gatewayPingbackHandler := func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		pingbackHandler(w, r)
	}
	handle(`GET`, `/v3/xmlrpc`, gatewayPingbackHandler)
	handle(`POST`, `/v3/xmlrpc`, gatewayPingbackHandler)

	handle(`GET`, `/v3/redirect-to-grafana`, redirectToGrafana)

	return nil
}

func redirectToGrafana(w http.ResponseWriter, req *http.Request, params map[string]string) {
	t := time.Now()
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix() * 1000
	end := time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.Local).Unix()*1000 - 1
	u := fmt.Sprintf(`/grafana/d/_2g5VpWnz/pei-ta-qu-liu-lang?orgId=1&from=%d&to=%d&refresh=5s&kiosk`, start, end)
	http.Redirect(w, req, u, http.StatusFound)
}

func (g *Gateway) GetAvatar(w http.ResponseWriter, req *http.Request, params map[string]string) {
	commentID, err := strconv.Atoi(params[`id`])
	if err != nil {
		panic(err)
	}

	in := &protocols.GetAvatarRequest{
		CommentID:       int64(commentID),
		IfModifiedSince: req.Header.Get("If-Modified-Since"),
		IfNoneMatch:     req.Header.Get("If-None-Match"),
		SetStatus: func(statusCode int) {
			w.WriteHeader(statusCode)
		},
		SetHeader: func(name string, value string) {
			w.Header().Add(name, value)
		},
		W: w,
	}
	g.service.GetAvatar(in)
}

func getAPI(w http.ResponseWriter, req *http.Request, params map[string]string) {
	http.ServeFile(w, req, `protocols/docs/index.html`)
}

func getSwagger(w http.ResponseWriter, req *http.Request, params map[string]string) {
	http.ServeFile(w, req, `protocols/docs/taoblog.swagger.json`)
}

func deprecated(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
}
