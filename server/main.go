package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/themes/blog"
	"github.com/movsb/taorm/taorm"
)

func main() {
	var err error

	cfg := config.LoadFile(`taoblog.yml`)

	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Endpoint,
		cfg.Database.Database,
	)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(10)
	defer db.Close()

	migration.Migrate(db)

	apiRouter := gin.Default()
	theAPI := apiRouter.Group("/v2")

	theAPI.Use(func(c *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				if iHTTPError, ok := e.(exception.IHTTPError); ok {
					err := iHTTPError.ToHTTPError()
					c.JSON(err.Code, err)
					return
				}
				if err, ok := e.(error); ok {
					if taorm.IsNotFoundError(err) {
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
	theService := service.NewService(cfg, db, theAuth)
	theAuth.SetLogin(theService.GetDefaultStringOption("login", "x"))
	theAuth.SetKey(cfg.Auth.Key)
	theAuth.SetGitHub(
		cfg.Auth.Github.ClientID,
		cfg.Auth.Github.ClientSecret,
		cfg.Auth.Github.UserID,
	)

	gateway.NewGateway(theAPI, theService, theAuth)

	var adminRouter *gin.Engine

	if !cfg.Maintenance.DisableAdmin {
		adminRouter = gin.Default()
		admin.NewAdmin(theService, theAuth, adminRouter.Group("/admin"))
	}

	themeRouter := gin.Default()
	indexGroup := themeRouter.Group("/", maybeSiteClosed(theService, theAuth))

	switch cfg.Theme.Name {
	case "", "BLOG":
		blog.NewBlog(cfg, theService, theAuth, indexGroup, theAPI, "themes/blog")
	default:
		panic("unknown theme: " + cfg.Theme.Name)
	}

	apiPrefix := regexp.MustCompile(`^/v\d+/`)
	adminPrefix := regexp.MustCompile(`^/admin/`)

	handler := func(w http.ResponseWriter, req *http.Request) {
		if apiPrefix.MatchString(req.URL.Path) {
			apiRouter.ServeHTTP(w, req)
			return
		}
		if adminRouter != nil && adminPrefix.MatchString(req.URL.Path) {
			adminRouter.ServeHTTP(w, req)
			return
		}
		themeRouter.ServeHTTP(w, req)
	}

	server := &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: http.HandlerFunc(handler),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGKILL)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit

	log.Println("server shutting down")
	server.Shutdown(context.Background())
	log.Println("server shutted down")
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
