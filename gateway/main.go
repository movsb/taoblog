package gateway

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
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
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true,
		}),
	)

	mux.Handle(`/v3/`, mux2)

	if err := g.runHTTPService(context.TODO(), mux, mux2, service); err != nil {
		panic(err)
	}

	return g
}

// runHTTPService ...
// TODO auth
func (g *Gateway) runHTTPService(ctx context.Context, mux *http.ServeMux, mux2 *runtime.ServeMux, svr protocols.TaoBlogServer) error {
	protocols.RegisterTaoBlogHandlerServer(ctx, mux2, svr)

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

	handle(`GET`, `/v3/avatars/{hash=*}`, g.GetAvatar)

	handle(`GET`, `/files/{post_id}/{file=**}`, g.GetFile)
	handle(`POST`, `/v3/posts/{post_id}/{file=**}`, g.CreateFile)
	handle(`DELETE`, `/v3/posts/{post_id}/files/{file=**}`, g.DeleteFile)

	return nil
}

func (g *Gateway) GetAvatar(w http.ResponseWriter, req *http.Request, params map[string]string) {
	query := req.URL.RawQuery
	in := &protocols.GetAvatarRequest{
		Query:           params[`hash`] + `?` + query,
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
