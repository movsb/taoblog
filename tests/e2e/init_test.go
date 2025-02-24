package e2e_test

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type R struct {
	server *server.Server
	client *clients.ProtoClient
	admin  context.Context
	guest  context.Context
	user1  context.Context
	user2  context.Context

	user1ID, user2ID int64
}

// 会在服务启动后快速返回。
func Serve(ctx context.Context, options ...server.With) *R {
	// 测试环境应该不依赖本地系统。
	version.EnableDevMode = false

	cfg := config.DefaultConfig()
	cfg.Database.Posts = "" // 使用内存
	cfg.Database.Files = "" // 使用内存
	cfg.Server.HTTPListen = `localhost:0`

	r := &R{}

	r.server = server.NewServer(options...)
	ready := make(chan struct{})
	go r.server.Serve(ctx, true, &cfg, ready)
	<-ready

	// 测试的时候默认禁用限流器；测试限流器相关函数会手动开启。
	r.server.TestEnableRequestThrottler(false)

	r.client = clients.NewProtoClientFromAddress(r.server.GRPCAddr())

	r.admin = onBehalfOf(r, int64(auth.AdminID))
	r.guest = context.Background()

	u := utils.Must1(r.server.Main().CreateUser(r.admin, &proto.User{Nickname: `用户1`}))
	r.user1 = onBehalfOf(r, u.Id)
	r.user1ID = u.Id
	u = utils.Must1(r.server.Main().CreateUser(r.admin, &proto.User{Nickname: `用户2`}))
	r.user2 = onBehalfOf(r, u.Id)
	r.user2ID = u.Id

	return r
}

func onBehalfOf(r *R, user int64) context.Context {
	u := &auth.User{User: utils.Must1(r.server.Auth().GetUserByID(user))}
	return auth.TestingUserContext(u, "go_test")
}
