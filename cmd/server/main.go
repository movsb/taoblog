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
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/metrics"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/theme"
	"github.com/movsb/taoblog/theme/modules/canonical"
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
		a := admin.NewAdmin(theService, theAuth, `/admin/`)
		log.Println(`admin on`, a.Prefix())
		mux.Handle(a.Prefix(), a.Handler())
	}

	theme := theme.New(cfg, theService, theAuth, `theme/blog`)
	canon := canonical.New(theme, r)
	mux.Handle(`/`, canon)
	theService.SetLinker(theme.Linker())

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

	log.Println("Server started on", server.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
	close(quit)

	log.Println("server shutting down")
	server.Shutdown(context.Background())
	log.Println("server shut down")
}

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
				migration.Init(cfg, db)
				return db
			}
		}
		panic(err)
	}
	return db
}
