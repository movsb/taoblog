package e2e_test

import (
	"context"
	"net/http"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
)

type R struct {
	server *server.Server

	// 无凭证的客户端。
	client *clients.ProtoClient

	system context.Context
	admin  context.Context
	guest  context.Context
	user1  context.Context
	user2  context.Context

	user1ID, user2ID int64
}

// 会在服务启动后快速返回。
func Serve(ctx context.Context, options ...server.With) *R {
	// 测试环境应该不依赖本地系统。
	version.ForceEnableDevMode = `0`

	cfg := config.DefaultConfig()
	cfg.Database.Posts = "" // 使用内存
	cfg.Database.Files = "" // 使用内存
	cfg.Database.Cache = "" // 使用内存
	cfg.Server.HTTPListen = `localhost:0`

	r := &R{}

	r.server = server.NewServer(options...)
	ready := make(chan struct{})
	go r.server.Serve(ctx, true, cfg, ready)
	<-ready

	// 测试的时候默认禁用限流器；测试限流器相关函数会手动开启。
	r.server.TestEnableRequestThrottler(false)

	r.client = clients.NewFromAddress(r.server.GRPCAddr(), ``)

	r.system = onBehalfOf(r, int64(user.SystemID))
	r.admin = onBehalfOf(r, int64(user.AdminID))
	r.guest = context.Background()

	u := utils.Must1(r.client.Users.CreateUser(r.admin, &proto.User{Nickname: `用户1`}))
	r.user1 = onBehalfOf(r, u.Id)
	r.user1ID = u.Id
	u = utils.Must1(r.client.Users.CreateUser(r.admin, &proto.User{Nickname: `用户2`}))
	r.user2 = onBehalfOf(r, u.Id)
	r.user2ID = u.Id

	return r
}

func onBehalfOf(r *R, uid int64) context.Context {
	u := utils.Must1(r.server.AuthFrontend().GetUserByID(int(uid)))
	return user.TestingUserContextForClient(u)
}

func (r *R) addAuth(req *http.Request, uid int64) {
	u := utils.Must1(r.server.Auth().GetUserByID(context.TODO(), int(uid)))
	req.Header.Set(`Authorization`, u.AuthorizationValue())
}
