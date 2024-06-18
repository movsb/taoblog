package gateway

import (
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/gateway/handlers/apidoc"
	"github.com/movsb/taoblog/gateway/handlers/assets"
	"github.com/movsb/taoblog/gateway/handlers/features"
	"github.com/movsb/taoblog/gateway/handlers/robots"
	"github.com/movsb/taoblog/gateway/handlers/rss"
	"github.com/movsb/taoblog/gateway/handlers/sitemap"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/go/handy"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/service/modules/webhooks"
	"github.com/movsb/taoblog/theme/modules/handle304"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

type Gateway struct {
	mux     *http.ServeMux
	service *service.Service
	auther  *auth.Auth

	instantNotifier notify.InstantNotifier
}

func NewGateway(service *service.Service, auther *auth.Auth, mux *http.ServeMux, instantNotifier notify.InstantNotifier) *Gateway {
	g := &Gateway{
		mux:     mux,
		service: service,
		auther:  auther,

		instantNotifier: instantNotifier,
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

	if err := g.register(context.TODO(), mux, mux2); err != nil {
		panic(err)
	}

	return g
}

func (g *Gateway) register(ctx context.Context, mux *http.ServeMux, mux2 *runtime.ServeMux) error {
	mc := utils.ServeMuxChain{ServeMux: mux}

	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100<<20),
			grpc.MaxCallSendMsgSize(100<<20),
		),
	}

	proto.RegisterUtilsHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), dialOptions)
	proto.RegisterTaoBlogHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), dialOptions)
	proto.RegisterSearchHandlerFromEndpoint(ctx, mux2, g.service.GrpcAddress(), dialOptions)

	mux2.HandlePath(`GET`, `/v3/avatar/{id}`, g.getAvatar)

	mux2.HandlePath(`POST`, `/v3/webhooks/github`, g.githubWebhook())
	mux2.HandlePath(`POST`, `/v3/webhooks/grafana/notify`, g.grafanaNotify)

	mux.Handle(`GET /v3/dynamic/`, http.StripPrefix(`/v3/dynamic`, &dynamic.Handler{}))

	info := utils.Must1(g.service.GetInfo(ctx, &proto.GetInfoRequest{}))

	// 博客功能集
	mc.Handle(`GET /v3/features/{theme}`, features.New())

	// API 文档
	mc.Handle(`GET /v3/api/`, http.StripPrefix(`/v3/api`, apidoc.New()))

	// 文件服务：/123/a.txt
	mc.Handle(`GET /v3/posts/{id}/files`, assets.New(`post`, g.service.GrpcAddress()), g.mimicGateway)

	// 站点地图：sitemap.xml
	mc.Handle(`GET /sitemap.xml`, sitemap.New(g.service, g.service), g.lastPostTimeHandler)

	// 订阅：rss
	mc.Handle(`GET /rss`, rss.New(g.service, rss.WithArticleCount(10)), g.lastPostTimeHandler)

	// 机器人控制：robots.txt
	sitemapFullURL := utils.Must1(url.Parse(info.Home)).JoinPath(`sitemap.xml`).String()
	mux.Handle(`GET /robots.txt`, robots.NewDefaults(sitemapFullURL))

	return nil
}

func (g *Gateway) lastPostTimeHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := utils.Must1(g.service.GetInfo(r.Context(), &proto.GetInfoRequest{}))
		handle304.New(h,
			handle304.WithNotModified(time.Unix(int64(info.LastPostedAt), 0)),
			handle304.WithEntityTag(version.GitCommit, version.Time, info.LastPostedAt),
		).ServeHTTP(w, r)
	})
}

func (g *Gateway) mimicGateway(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		md := metadata.Pairs()
		for _, cookie := range r.Header.Values(`cookie`) {
			md.Append(auth.GatewayCookie, cookie)
		}
		for _, userAgent := range r.Header.Values(`user-agent`) {
			md.Append(auth.GatewayUserAgent, userAgent)
		}
		ctx := metadata.NewOutgoingContext(r.Context(), md)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *Gateway) githubWebhook() runtime.HandlerFunc {
	h := webhooks.CreateHandler(
		g.service.Config().Maintenance.Webhook.GitHub.Secret,
		g.service.Config().Maintenance.Webhook.ReloaderPath,
		func(content string) {
			g.instantNotifier.InstantNotify("博客更新", content)
		},
	)
	return func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		h(w, req)
	}
}

func (g *Gateway) grafanaNotify(w http.ResponseWriter, req *http.Request, params map[string]string) {
	body := utils.DropLast1(io.ReadAll(io.LimitReader(req.Body, 1<<20)))
	var m map[string]any
	json.Unmarshal(body, &m)
	var message string
	if x, ok := m[`message`]; ok {
		message, _ = x.(string)
	}
	g.service.UtilsServer.InstantNotify(g.auther.NewContextForRequest(req), &proto.InstantNotifyRequest{
		Title: `监控告警`,
		// https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/
		Message: message,
	})
}

func (g *Gateway) getAvatar(w http.ResponseWriter, req *http.Request, params map[string]string) {
	ephemeral, err := strconv.Atoi(params[`id`])
	if err != nil {
		panic(err)
	}
	in := &handy.GetAvatarRequest{
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
