package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/themes/front"
	"github.com/movsb/taoblog/themes/weekly"
)

func main() {
	var err error

	dataSource := fmt.Sprintf("%s:%s@/%s",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
	)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(10)
	defer db.Close()

	migration.Migrate(db)

	router := gin.Default()

	theAPI := router.Group("/v2")

	theAPI.Use(func(c *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				if iHTTPError, ok := e.(exception.IHTTPError); ok {
					err := iHTTPError.ToHTTPError()
					c.JSON(err.Code, err)
					return
				}
				if err, ok := e.(error); ok {
					if err == sql.ErrNoRows {
						c.Status(404)
						return
					}
				}
				panic(e)
			}
		}()
		c.Next()
	})

	theAPI.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	theAuth := &auth.Auth{}
	theService := service.NewService(db, theAuth)
	gateway.NewGateway(theAPI, theService, theAuth)
	admin.NewAdmin(theService, theAuth, router.Group("/admin"))

	switch themeName := os.Getenv("THEME"); themeName {
	case "BLOG":
		front.NewFront(theService, theAuth, router.Group("/blog"), theAPI, "themes/front")
	case "WEEKLY":
		weekly.NewWeekly(theService, theAuth, router.Group("/blog"), theAPI, "themes/weekly")
	default:
		panic("unknown theme")
	}

	theAuth.SetLogin(theService.GetDefaultStringOption("login", "x"))
	theAuth.SetKey(os.Getenv("KEY"))

	server := &http.Server{
		Addr:    os.Getenv("LISTEN"),
		Handler: router,
	}

	go server.ListenAndServe()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGKILL)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit

	log.Println("server shutting down")
	server.Shutdown(context.Background())
	log.Println("server shutted down")
}
