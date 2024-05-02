package migration

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/service/models"
	setup_data "github.com/movsb/taoblog/setup/data"
	"github.com/movsb/taorm"
)

// Init ...
func Init(cfg *config.Config, db *sql.DB) {
	var err error
	var cmd *exec.Cmd

	var fp io.ReadCloser
	defer func() {
		if fp != nil {
			fp.Close()
		}
	}()

	switch cfg.Database.Engine {
	case `sqlite`:
		cmd = exec.Command(`sqlite3`, cfg.Database.SQLite.Path)
		fp, err = setup_data.Root.Open(`schemas.sqlite.sql`)
		if err != nil {
			panic(err)
		}
	default:
		panic("unknown database engine")
	}

	cmd.Stdin = fp
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		log.Fatalln(err)
	}

	tdb := taorm.NewDB(db)
	tdb.MustTxCall(func(tx *taorm.DB) {
		now := int32(time.Now().Unix())
		tx.Model(&models.Option{
			Name:  `db_ver`,
			Value: fmt.Sprint(MaxVersionNumber()),
		}).MustCreate()
		tx.Model(&models.Post{
			Date:       now,
			Modified:   now,
			Title:      `你好，世界`,
			Type:       `post`,
			Category:   0,
			Status:     `public`,
			SourceType: `markdown`,
			Source:     `你好，世界！这是您的第一篇文章。`,
		}).MustCreate()
	})
}
