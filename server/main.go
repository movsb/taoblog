package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/movsb/taoblog/modules/canonical"

	_ "github.com/mattn/go-sqlite3" // shut up
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/auth"
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
	cfg := config.LoadFile(`taoblog.yml`)

	db := initDatabase(cfg)
	db.SetMaxIdleConns(10)
	defer db.Close()

	migration.Migrate(db)

	theAuth := auth.New(cfg.Auth)
	theService := service.NewService(cfg, db, theAuth)

	var mux = http.NewServeMux()

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

	canon := canonical.New(renderer)
	mux.Handle(`/`, canon)

	server := &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: mux,
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
