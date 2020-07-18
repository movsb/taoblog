package inits

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taorm/taorm"
)

// Init ...
func Init(cfg *config.Config, db *sql.DB) {
	var err error
	var cmd *exec.Cmd
	var fp io.ReadCloser
	switch cfg.Database.Engine {
	case `sqlite`:
		cmd = exec.Command(`sqlite3`, cfg.Database.SQLite.Path)
		fp, err = os.Open(`setup/data/schemas.sqlite.sql`)
		if err != nil {
			panic(err)
		}
	default:
		panic("unknown database engine")
	}
	defer fp.Close()
	cmd.Stdin = fp
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		panic(err)
	}

	tdb := taorm.NewDB(db)
	tdb.MustTxCall(func(tx *taorm.DB) {
		now := int32(time.Now().Unix())
		tx.Model(&models.Option{
			Name:  `db_ver`,
			Value: fmt.Sprint(migration.MaxVersionNumber()),
		}).MustCreate()
		tx.Model(&models.Post{
			Date:       now,
			Modified:   now,
			Title:      `你好，世界`,
			Content:    `你好，世界！这是您的第一篇文章。`,
			Type:       `post`,
			Category:   0,
			Status:     `public`,
			SourceType: `markdown`,
			Source:     `你好，世界！这是您的第一篇文章。`,
		}).MustCreate()
	})
}
