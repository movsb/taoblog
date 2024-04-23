package gateway

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/pkg/notify"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	proto_docs "github.com/movsb/taoblog/protocols/docs"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/modules/webhooks"
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

	mux.Handle(`/v3/`, mux2)

	if err := g.register(context.TODO(), mux2); err != nil {
		panic(err)
	}

	return g
}

func (g *Gateway) register(ctx context.Context, mux2 *runtime.ServeMux) error {
	protocols.RegisterTaoBlogHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), []grpc.DialOption{grpc.WithInsecure()})
	protocols.RegisterSearchHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), []grpc.DialOption{grpc.WithInsecure()})

	mux2.HandlePath("GET", `/v3/api`, serveProtoDocsFile(`index.html`))
	mux2.HandlePath("GET", `/v3/api/swagger`, serveProtoDocsFile(`taoblog.swagger.json`))

	mux2.HandlePath(`GET`, `/v3/avatar/{id}`, g.getAvatar)

	mux2.HandlePath(`POST`, `/v3/webhooks/github`, g.githubWebhook())

	return nil
}

func (g *Gateway) githubWebhook() func(w http.ResponseWriter, req *http.Request, params map[string]string) {
	h := webhooks.CreateHandler(
		g.service.Config().Maintenance.Webhook.GitHub.Secret,
		g.service.Config().Maintenance.Webhook.ReloaderPath,
		func(content string) {
			tk := g.service.Config().Comment.Push.Chanify.Token
			ch := notify.NewOfficialChanify(tk)
			ch.Send("博客更新", content, true)
		},
	)
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		h(w, req)
	}
}

func (g *Gateway) getAvatar(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ephemeral, err := strconv.Atoi(params[`id`])
	if err != nil {
		panic(err)
	}
	in := &protocols.GetAvatarRequest{
		Ephemeral:       ephemeral,
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

func serveProtoDocsFile(path string) func(w http.ResponseWriter, req *http.Request, params map[string]string) {
	_, name := filepath.Split(path)
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		fp, err := proto_docs.Root.Open(path)
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		stat, err := fp.Stat()
		if err != nil {
			panic(err)
		}
		rs, ok := fp.(io.ReadSeeker)
		if !ok {
			panic(`bad embed file`)
		}
		http.ServeContent(w, req, name, stat.ModTime(), rs)
	}
}
