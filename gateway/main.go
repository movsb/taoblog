package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	_ "embed"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/theme/modules/handle304"
)

type Gateway struct {
	mux     *http.ServeMux
	mc      *utils.ServeMuxChain
	service *service.Service

	// 未鉴权的
	client *clients.ProtoClient

	notify proto.NotifyServer
}

func NewGateway(serverAddr string, service *service.Service, mux *http.ServeMux, notify proto.NotifyServer) *Gateway {
	g := &Gateway{
		mux:     mux,
		mc:      &utils.ServeMuxChain{ServeMux: mux},
		service: service,

		client: clients.NewFromAddress(serverAddr, ``),

		notify: notify,
	}

	if err := g.register(context.Background(), serverAddr, mux); err != nil {
		panic(err)
	}

	return g
}

// 网站头像
func (g *Gateway) SetFavicon(f *favicon.Favicon) {
	g.mux.Handle(`/favicon.ico`, f)
}

// 扩展功能动态生成的样式、脚本、文件。
func (g *Gateway) SetDynamic(invalidate func()) {
	g.mux.Handle(dynamic.PrefixSlashed, http.StripPrefix(dynamic.Prefix, dynamic.New(invalidate)))
}

// 订阅：rss
func (g *Gateway) SetRSS(rss *rss.RSS) {
	g.mc.Handle(`GET /rss`, rss, g.lastPostTimeHandler)
}

func (g *Gateway) AvatarURL(uid int) string {
	e := g.service.UserAvatarEphemeral(context.Background(), uid, "")
	return fmt.Sprintf(`/v3/avatar/%d`, e)
}

// 头像服务
func (g *Gateway) SetAvatar(ctx context.Context, cache *cache.FileCache, resolve avatar.ResolveFunc) {
	task := avatar.NewTask(ctx, cache)
	a := avatar.New(task, resolve)
	g.mc.Handle(`GET /v3/avatar/{id}`, a.Handler())
}

// 文章 Open Graph 背景图。
func (g *Gateway) SetOpenGraphImageHandler(h http.HandlerFunc) {
	g.mc.HandleFunc(`GET /v3/posts/{id}/open_graph.png`, h)
}

func (g *Gateway) register(ctx context.Context, serverAddr string, mux *http.ServeMux) error {
	mc := utils.ServeMuxChain{ServeMux: mux}

	info := utils.Must1(g.client.Blog.GetInfo(ctx, &proto.GetInfoRequest{}))

	// 无需鉴权的部分
	// 可跨进程使用。
	{
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
		mc.Handle(`POST /v3/webhooks/grafana/notify`, grafana.New(g.client.Notify))

		// GRPC 走 HTTP 通信。少暴露一个端口，降低架构复杂性。
		mc.Handle(`GET /v3/grpc`, grpc_proxy.New(serverAddr))
	}

	// 使用登录身份鉴权的部分
	// 可跨进程使用。
	{
		api := api.New(ctx, g.client,
			func(mux *runtime.ServeMux) {
				// /v3/posts/.../files 被 GRPC 注册了，但是上层由于注册的是 /v3/，所以 GRPC 那里
				// 接收不到这个不带 / 结尾的请求。会被重定向到 files/。
				// 所以这里注册到 api 里面。
				// https://github.com/grpc-ecosystem/grpc-gateway/issues/5771
				mux.HandlePath(`GET`, `/v3/posts/{id}/files/{path=**}`, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
					pid := utils.Must1(strconv.Atoi(pathParams[`id`]))
					path := pathParams[`path`]
					if pid <= 0 || path == `` {
						utils.HTTPError(w, 400)
						return
					}
					assets.GetFile(g.service, w, r, pid, path)
				})
			},
			nil,
		)

		// GRPC API 转接层
		mc.Handle(`/v3/`, api)

		// 文件服务
		mc.Handle(`POST /v3/posts/{id}/files`, assets.CreateFile(g.client))

		// 站点地图：sitemap.xml
		mc.Handle(`GET /sitemap.xml`, sitemap.New(g.client, g.service), g.lastPostTimeHandler)

		// 调试相关
		mc.Handle(`/debug/{session}/`, debug.Handler())
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

type AuthFunc func(g *Gateway, r *http.Request) *user.User
