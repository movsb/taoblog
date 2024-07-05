package e2e_test

import (
	"context"
	"os"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/clients"
)

var Server *server.Server
var client clients.Client
var admin context.Context
var guest context.Context

// 会在服务启动后快速返回。
func Serve(ctx context.Context) {
	// 测试环境应该不依赖本地系统。
	os.Setenv(`DEV`, `0`)

	cfg := config.DefaultConfig()
	cfg.Auth.Basic.Username = `test`
	cfg.Auth.Basic.Password = `test`
	cfg.Auth.Key = `12345678`
	cfg.Database.Path = ""  // 使用内存
	cfg.Data.File.Path = "" // 使用内存
	cfg.Server.HTTPListen = `localhost:0`
	cfg.Server.GRPCListen = `localhost:0`

	Server = &server.Server{}
	ready := make(chan struct{})
	go Server.Serve(ctx, true, &cfg, ready)
	<-ready

	// 测试的时候默认禁用限流器；测试限流器相关函数会手动开启。
	Server.Service.TestEnableRequestThrottler(false)

	client = clients.NewFromGrpcAddr(Server.GRPCAddr)
	admin = auth.TestingAdminUserContext(Server.Auther, "go_test")
	guest = context.Background()
}

func init() {
	Serve(context.Background())
}
