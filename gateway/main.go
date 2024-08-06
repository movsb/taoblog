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
	"github.com/movsb/taoblog/gateway/handlers/debug"
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

	client          clients.Client
	instantNotifier notify.InstantNotifier
}

func NewGateway(service *service.Service, auther *auth.Auth, mux *http.ServeMux, instantNotifier notify.InstantNotifier) *Gateway {
	g := &Gateway{
		mux:     mux,
		service: service,
		auther:  auther,

		client:          clients.NewFromGrpcAddr(service.Addr().String()),
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
		mc.Handle(`/v3/webhooks/github/`, github.New(g.client,
			g.service.Config().Maintenance.Webhook.GitHub.Secret,
			g.instantNotifier, `/v3/webhooks/github`,
		))

		// Grafana 监控告警通知。
		mc.Handle(`POST /v3/webhooks/grafana/notify`, grafana.New(g.auther, g.client))
	}

	// 使用系统帐号鉴权的部分
	// 只能在进程内使用。
	{
		// 头像服务
		task := avatar.NewTask(g.service.GetPluginStorage(`avatar`))
		mc.Handle(`GET /v3/avatar/{id}`, avatar.New(task, g.service))
	}

	// 使用登录身份鉴权的部分
	// 可跨进程使用。
	{
		// GRPC API 转接层
		mc.Handle(`/v3/`, api.New(ctx, g.client))

		// 文件服务：/123/a.txt
		mc.Handle(`GET /v3/posts/{id}/files`, assets.New(g.auther, `post`, g.client))

		// 站点地图：sitemap.xml
		mc.Handle(`GET /sitemap.xml`, sitemap.New(g.auther, g.client, g.service), g.lastPostTimeHandler)

		// 订阅：rss
		mc.Handle(`GET /rss`, rss.New(g.auther, g.client, rss.WithArticleCount(10)), g.lastPostTimeHandler)

		// 调试相关
		mc.Handle(`/debug/`, http.StripPrefix(`/debug`, debug.Handler()), g.adminHandler)
	}

	return nil
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

func (g *Gateway) adminHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac := auth.Context(r.Context())
		if !ac.User.IsAdmin() {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}
