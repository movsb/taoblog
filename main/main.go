package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/front"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/service"
)

const (
	dbVer = "4"
)

func checkDBVer(db *sql.DB) {
	ver := ""
	s := `SELECT value FROM options WHERE name='db_ver'`
	row := db.QueryRow(s)
	if err := row.Scan(&ver); err != nil {
		panic(fmt.Sprintf("cannot check db_ver: %v", err))
	}
	if ver != dbVer {
		panic(fmt.Sprintf("db_ver mismatch: %s != %s", ver, dbVer))
	}
}

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

	checkDBVer(db)

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
	theService := service.NewImplServer(db, theAuth)
	gateway.NewGateway(theAPI, theService, theAuth)
	front.NewFront(theService, theAuth, router.Group("/blog"), theAPI)

	theAuth.SetLogin(theService.GetDefaultStringOption("login", "x"))
	theAuth.SetKey(os.Getenv("KEY"))

	server := &http.Server{
		Addr:    os.Getenv("LISTEN"),
		Handler: router,
	}

	go server.ListenAndServe()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("server shutting down")
	server.Shutdown(context.Background())
	log.Println("server shutted down")
}
