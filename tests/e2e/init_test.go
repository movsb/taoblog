package e2e_test

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
)

var Server *server.Server
var client *clients.ProtoClient
var admin context.Context
var guest context.Context

// 会在服务启动后快速返回。
func Serve(ctx context.Context) {
	// 测试环境应该不依赖本地系统。
	version.EnableDevMode = false

	cfg := config.DefaultConfig()
	cfg.Database.Path = ""  // 使用内存
	cfg.Data.File.Path = "" // 使用内存
	cfg.Server.HTTPListen = `localhost:0`

	Server = server.NewDefaultServer()
	ready := make(chan struct{})
	go Server.Serve(ctx, true, &cfg, ready)
	<-ready

	// 测试的时候默认禁用限流器；测试限流器相关函数会手动开启。
	Server.TestEnableRequestThrottler(false)

	client = clients.NewProtoClientFromAddress(Server.GRPCAddr())
	adminUser := &auth.User{User: utils.Must1(Server.Auth().GetUserByID(int64(auth.AdminID)))}
	admin = auth.TestingUserContext(adminUser, "go_test")
	guest = context.Background()
}

func init() {
	Serve(context.Background())
}
