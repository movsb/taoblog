package service_test

import (
	"context"
	"log"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/setup/migration"
)

var blog proto.TaoBlogClient
var admin context.Context
var guest context.Context

func initService() {
	cfg := config.DefaultConfig()
	cfg.Auth.Basic.Username = `test`
	cfg.Auth.Basic.Password = `test`
	cfg.Database.SQLite.Path = "" // 使用内存
	cfg.Server.HTTPListen = `localhost:0`
	cfg.Server.GRPCListen = `localhost:0`

	db := server.InitDatabase(`sqlite3`, ``)
	// defer db.Close()

	migration.Migrate(db)

	log.Println(`DevMode:`, service.DevMode())

	theAuth := auth.New(cfg.Auth, service.DevMode())
	theService := service.NewService(&cfg, db, theAuth)
	theAuth.SetService(theService)

	blog = proto.NewTaoBlogClient(proto.NewConn("", theService.Addr().String()))
	admin = auth.TestingAdminUserContext(theAuth, "go_test")
	guest = context.Background()
}
