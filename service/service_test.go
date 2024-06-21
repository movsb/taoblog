package service_test

import (
	"context"
	"log"
	"os"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/setup/migration"
)

var blog proto.TaoBlogClient
var admin context.Context
var guest context.Context

func initService() {
	// 测试环境应该不依赖本地系统。
	os.Setenv(`DEV`, `0`)

	cfg := config.DefaultConfig()
	cfg.Auth.Basic.Username = `test`
	cfg.Auth.Basic.Password = `test`
	cfg.Database.Path = "" // 使用内存
	cfg.Server.HTTPListen = `localhost:0`
	cfg.Server.GRPCListen = `localhost:0`

	db := server.InitDatabase(``)
	// defer db.Close()

	migration.Migrate(db)

	log.Println(`DevMode:`, service.DevMode())

	theAuth := auth.New(cfg.Auth, service.DevMode())
	theService := service.NewServiceForTesting(&cfg, db, theAuth)
	theAuth.SetService(theService)

	blog = proto.NewTaoBlogClient(clients.NewConn("", theService.Addr().String()))
	admin = auth.TestingAdminUserContext(theAuth, "go_test")
	guest = context.Background()
}
