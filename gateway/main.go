package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	_ "embed"

	"github.com/movsb/taoblog/gateway/addons"
	"github.com/movsb/taoblog/gateway/handlers/api"
	"github.com/movsb/taoblog/gateway/handlers/apidoc"
	"github.com/movsb/taoblog/gateway/handlers/assets"
	"github.com/movsb/taoblog/gateway/handlers/avatar"
	"github.com/movsb/taoblog/gateway/handlers/debug"
	"github.com/movsb/taoblog/gateway/handlers/favicon"
	"github.com/movsb/taoblog/gateway/handlers/features"
	grpc_proxy "github.com/movsb/taoblog/gateway/handlers/grpc"
	"github.com/movsb/taoblog/gateway/handlers/robots"
	"github.com/movsb/taoblog/gateway/handlers/rss"
	"github.com/movsb/taoblog/gateway/handlers/sitemap"
	"github.com/movsb/taoblog/gateway/handlers/webhooks/github"
	"github.com/movsb/taoblog/gateway/handlers/webhooks/grafana"
	"github.com/movsb/taoblog/modules/auth"
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

	// 未鉴权的
	client *clients.ProtoClient

	notify proto.NotifyServer
}

func NewGateway(serverAddr string, service *service.Service, auther *auth.Auth, mux *http.ServeMux, notify proto.NotifyServer) *Gateway {
	g := &Gateway{
		mux:     mux,
		service: service,
		auther:  auther,

		client: clients.NewProtoClientFromAddress(serverAddr),

		notify: notify,
	}

	if err := g.register(context.TODO(), serverAddr, mux); err != nil {
		panic(err)
	}

	return g
}

func (g *Gateway) SetFavicon(f *favicon.Favicon) {
	// 网站头像
	g.mux.Handle(`/favicon.ico`, f)
}

func (g *Gateway) register(ctx context.Context, serverAddr string, mux *http.ServeMux) error {
	mc := utils.ServeMuxChain{ServeMux: mux}

	info := utils.Must1(g.client.Blog.GetInfo(ctx, &proto.GetInfoRequest{}))

	// 无需鉴权的部分
	// 可跨进程使用。
	{
		// 网站头像
		// mc.Handle(`/favicon.ico`, favicon.NewFavicon())

		// 扩展功能动态生成的样式、脚本、文件。
		mc.Handle(dynamic.PrefixSlashed, http.StripPrefix(dynamic.Prefix, dynamic.New()))

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
			g.notify, `/v3/webhooks/github`,
		))

		// Grafana 监控告警通知。
		mc.Handle(`POST /v3/webhooks/grafana/notify`, grafana.New(g.auther, g.client.Notify))

		// GRPC 走 HTTP 通信。少暴露一个端口，降低架构复杂性。
		mc.Handle(`GET /v3/grpc`, grpc_proxy.New(serverAddr))
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
		mc.Handle(`/debug/`, http.StripPrefix(`/debug`, debug.Handler()), g.nonGuestHandler(userFromRequestContext, userFromRequestQuery))

		// 附加组件提供的
		mc.Handle(`/v3/addons/`, http.StripPrefix(`/v3/addons`, addons.Handler()), g.nonGuestHandler(userFromRequestContext, userFromRequestQuery))
	}

	return nil
}

func (g *Gateway) lastPostTimeHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := utils.Must1(g.client.Blog.GetInfo(r.Context(), &proto.GetInfoRequest{}))
		handle304.New(h,
			handle304.WithNotModified(time.Unix(int64(info.LastPostedAt), 0)),
			handle304.WithEntityTag(version.GitCommit, version.Time, info.LastPostedAt),
		).ServeHTTP(w, r)
	})
}

type AuthFunc func(g *Gateway, r *http.Request) *auth.User

func userFromRequestContext(g *Gateway, r *http.Request) *auth.User {
	return auth.Context(r.Context()).User
}
func userFromRequestQuery(g *Gateway, r *http.Request) *auth.User {
	id, token, _ := auth.ParseAuthorizationValue(r.URL.Query().Get(`auth`))
	return g.auther.AuthLogin(fmt.Sprint(id), token)
}

func (g *Gateway) nonGuestHandler(fns ...AuthFunc) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var u *auth.User
			for _, fn := range fns {
				u = fn(g, r)
				if !u.IsGuest() {
					break
				}
			}
			if u != nil && !u.IsGuest() {
				h.ServeHTTP(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		})
	}
}
