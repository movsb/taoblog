package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mattn/go-sqlite3"
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/metrics"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/canonical"
	"github.com/movsb/taoblog/service"
	inits "github.com/movsb/taoblog/setup/init"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/themes/blog"
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.LoadFile(`taoblog.yml`)

	db := initDatabase(cfg)
	defer db.Close()

	migration.Migrate(db)

	var mux = http.NewServeMux()
	r := metrics.NewRegistry(context.TODO())
	mux.Handle(`/v3/metrics`, r.Handler()) // TODO: insecure

	theAuth := auth.New(cfg.Auth)
	theService := service.NewService(cfg, db, theAuth)

	gateway.NewGateway(theService, theAuth, mux)

	if !cfg.Maintenance.DisableAdmin {
		admin.NewAdmin(theService, theAuth, mux)
	}

	var renderer canonical.Renderer
	switch strings.ToLower(cfg.Theme.Name) {
	case "", "blog":
		renderer = blog.NewBlog(cfg, theService, theAuth, "themes/blog")
	default:
		panic("unknown theme: " + cfg.Theme.Name)
	}

	canon := canonical.New(renderer, r)
	mux.Handle(`/`, canon)

	server := &http.Server{
		Addr:    cfg.Server.HTTPListen,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	log.Println("Server started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGKILL)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
	close(quit)

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

func initDatabase(cfg *config.Config) *sql.DB {
	var db *sql.DB
	var err error

	switch cfg.Database.Engine {
	case `sqlite`:
		v := url.Values{}
		v.Set(`cache`, `shared`)
		v.Set(`mode`, `rwc`)
		u := url.URL{
			Scheme:   `file`,
			Opaque:   url.PathEscape(cfg.Database.SQLite.Path),
			RawQuery: v.Encode(),
		}
		db, err = sql.Open(`sqlite3`, u.String())
		if err == nil {
			db.SetMaxOpenConns(1)
		}
	default:
		panic(`unknown database engine`)
	}
	if err != nil {
		panic(err)
	}

	var count int
	row := db.QueryRow(`select count(1) from options`)
	if err := row.Scan(&count); err != nil {
		if se, ok := err.(sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				inits.Init(cfg, db)
				return db
			}
		}
		panic(err)
	}
	return db
}
