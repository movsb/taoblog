package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
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
	"github.com/movsb/taoblog/themes/blog"
	"github.com/movsb/taoblog/themes/weekly"
)

const serverPidFilename = "server.pid"

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
	theAuth.SetLogin(theService.GetDefaultStringOption("login", "x"))
	theAuth.SetKey(os.Getenv("KEY"))
	theAuth.SetGitHub(
		os.Getenv("GITHUB_CLIENT_ID"),
		os.Getenv("GITHUB_CLIENT_SECRET"),
		os.Getenv("GITHUB_ID"),
	)

	gateway.NewGateway(theAPI, theService, theAuth)

	if disableAdmin := os.Getenv("DISABLE_ADMIN"); disableAdmin != "1" {
		admin.NewAdmin(theService, theAuth, router.Group("/admin"))
	}

	indexGroup := router.Group("/blog", maybeSiteClosed(theService, theAuth))

	switch themeName := os.Getenv("THEME"); themeName {
	case "BLOG":
		blog.NewBlog(theService, theAuth, indexGroup, theAPI, "themes/blog")
	case "WEEKLY":
		weekly.NewWeekly(theService, theAuth, indexGroup, theAPI, "themes/weekly")
	default:
		panic("unknown theme")
	}

	server := &http.Server{
		Addr:    os.Getenv("LISTEN"),
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	ioutil.WriteFile(serverPidFilename, []byte(fmt.Sprint(os.Getpid())), 0644)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGKILL)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit

	log.Println("server shutting down")
	server.Shutdown(context.Background())
	log.Println("server shutted down")

	os.Remove(serverPidFilename)
}

const siteClosedTemplate = `<!doctype html>
<html>
<body>
<center><h1>503 Site Closed</h1></center>
<hr/>
</body>
</html>
`

func maybeSiteClosed(svc *service.Service, auther *auth.Auth) func(c *gin.Context) {
	return func(c *gin.Context) {
		shouldAbort := false
		if svc.IsSiteClosed() {
			user := auther.AuthHeader(c)
			if user.IsGuest() {
				user = auther.AuthCookie(c)
			}
			if user.IsGuest() {
				shouldAbort = true
			}
		}
		if shouldAbort {
			c.Status(503)
			c.Header("Retry-After", "86400")
			c.Writer.WriteString(siteClosedTemplate)
			c.Abort()
		} else {
			c.Next()
		}
	}
}
