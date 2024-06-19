package gateway

import (
	"context"
	"net/http"
	"net/url"
	"time"

	_ "embed"

	"github.com/movsb/taoblog/gateway/handlers/api"
	"github.com/movsb/taoblog/gateway/handlers/apidoc"
	"github.com/movsb/taoblog/gateway/handlers/assets"
	"github.com/movsb/taoblog/gateway/handlers/avatar"
	"github.com/movsb/taoblog/gateway/handlers/features"
	"github.com/movsb/taoblog/gateway/handlers/robots"
	"github.com/movsb/taoblog/gateway/handlers/rss"
	"github.com/movsb/taoblog/gateway/handlers/sitemap"
	"github.com/movsb/taoblog/gateway/handlers/webhooks/github"
	"github.com/movsb/taoblog/gateway/handlers/webhooks/grafana"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/theme/modules/handle304"

	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
)

type Gateway struct {
	mux     *http.ServeMux
	service *service.Service
	auther  *auth.Auth

	client clients.Client
	server _Server

	instantNotifier notify.InstantNotifier
}

type _Server interface {
	proto.UtilsServer
	proto.TaoBlogServer
}

func NewGateway(service *service.Service, auther *auth.Auth, mux *http.ServeMux, instantNotifier notify.InstantNotifier) *Gateway {
	g := &Gateway{
		mux:     mux,
		service: service,
		auther:  auther,

		client: clients.NewFromGrpcAddr(service.GrpcAddress()),
		server: service,

		instantNotifier: instantNotifier,
	}

	if err := g.register(context.TODO(), mux); err != nil {
		panic(err)
	}

	return g
}

func (g *Gateway) register(ctx context.Context, mux *http.ServeMux) error {
	mc := utils.ServeMuxChain{ServeMux: mux}

	info := utils.Must1(g.client.GetInfo(ctx, &proto.GetInfoRequest{}))

	// 无需鉴权的部分
	// 可跨进程使用。
	{
		// 扩展功能动态生成的样式、脚本、文件。
		mc.Handle(`GET /v3/dynamic/`, http.StripPrefix(`/v3/dynamic`, &dynamic.Handler{}))

		// 博客功能集
		mc.Handle(`GET /v3/features/{theme}`, features.New())

		// API 文档
		mc.Handle(`GET /v3/api/`, http.StripPrefix(`/v3/api`, apidoc.New()))

		// 机器人控制：robots.txt
		sitemapFullURL := utils.Must1(url.Parse(info.Home)).JoinPath(`sitemap.xml`).String()
		mux.Handle(`GET /robots.txt`, robots.NewDefaults(sitemapFullURL))
	}

	// 无需鉴权但本身自带鉴权的部分
	{
		// GitHub Workflow 完成回调。
		mc.Handle(`POST /v3/webhooks/github`, github.New(
			g.service.Config().Maintenance.Webhook.GitHub.Secret,
			g.service.Config().Maintenance.Webhook.ReloaderPath,
			func(content string) {
				g.instantNotifier.InstantNotify(`GitHub Webhooks`, content)
			},
		))

		// Grafana 监控告警通知。
		mc.Handle(`POST /v3/webhooks/grafana/notify`, grafana.New(g.server), g.localAuthenticate)
	}

	// 使用系统帐号鉴权的部分
	// 只能在进程内使用。
	{
		// 头像服务
		mc.Handle(`GET /v3/avatar/{id}`, avatar.New(g.server), g.systemAdmin)
	}

	// 使用登录身份鉴权的部分
	// 可跨进程使用。
	{
		// GRPC API 转接层
		mc.Handle(`/v3/`, api.New(ctx, g.client))

		// 文件服务：/123/a.txt
		mc.Handle(`GET /v3/posts/{id}/files`, assets.New(`post`, g.client))

		// 站点地图：sitemap.xml
		mc.Handle(`GET /sitemap.xml`, sitemap.New(g.service, g.service), g.lastPostTimeHandler)

		// 订阅：rss
		mc.Handle(`GET /rss`, rss.New(g.client, rss.WithArticleCount(10)), g.lastPostTimeHandler)
	}

	return nil
}

func (g *Gateway) localAuthenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(g.auther.NewContextForRequest(r))
		h.ServeHTTP(w, r)
	})
}

func (g *Gateway) systemAdmin(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.SystemAdmin(r.Context())
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *Gateway) lastPostTimeHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := utils.Must1(g.client.GetInfo(r.Context(), &proto.GetInfoRequest{}))
		handle304.New(h,
			handle304.WithNotModified(time.Unix(int64(info.LastPostedAt), 0)),
			handle304.WithEntityTag(version.GitCommit, version.Time, info.LastPostedAt),
		).ServeHTTP(w, r)
	})
}
