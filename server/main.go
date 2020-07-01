package server

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
	_ "github.com/go-sql-driver/mysql" // shut up

	//_ "github.com/mattn/go-sqlite3"    // shut up
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/exception"
	"github.com/movsb/taoblog/service"
	inits "github.com/movsb/taoblog/setup/init"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/themes/blog"
	"github.com/movsb/taorm/taorm"
	"github.com/spf13/cobra"
)

// AddCommands ...
func AddCommands(rootCmd *cobra.Command) {
	serveCommand := &cobra.Command{
		Use:   `server`,
		Short: `Run the server`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			serve()
		},
	}

	rootCmd.AddCommand(serveCommand)
}

func serve() {
	cfg := config.LoadFile(`taoblog.yml`)

	db := initDatabase(cfg)
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

	theAuth := auth.New(cfg.Auth)
	theService := service.NewService(cfg, db, theAuth)
	gateway.NewGateway(theAPI, theService, theAuth, apiRouter)

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

func initDatabase(cfg *config.Config) *sql.DB {
	var db *sql.DB
	var err error

	switch cfg.Database.Engine {
	case `mysql`:
		dataSource := fmt.Sprintf(`%s:%s@tcp(%s)/%s`,
			cfg.Database.MySQL.Username,
			cfg.Database.MySQL.Password,
			cfg.Database.MySQL.Endpoint,
			cfg.Database.MySQL.Database,
		)
		db, err = sql.Open(`mysql`, dataSource)
	case `sqlite`:
		db, err = sql.Open(`sqlite3`, cfg.Database.SQLite.Path)
	default:
		panic(`unknown database engine`)
	}
	if err != nil {
		panic(err)
	}

	var count int
	row := db.QueryRow(`select count(1) from options`)
	if err := row.Scan(&count); err != nil {
		inits.Init(cfg, db)
	}
	return db
}
