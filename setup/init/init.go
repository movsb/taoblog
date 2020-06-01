package inits

import (
	"database/sql"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm/taorm"
)

const dbVer = 14

// Init ...
func Init(cfg *config.Config, db *sql.DB) {
	var err error
	var cmd *exec.Cmd
	var fp io.ReadCloser
	switch cfg.Database.Engine {
	case `mysql`:
		host, port, _ := net.SplitHostPort(cfg.Database.MySQL.Endpoint)
		cmd = exec.Command(`mysql`,
			`-h`, host,
			`-P`, port,
			`-u`, cfg.Database.MySQL.Username,
			`--password=`+cfg.Database.MySQL.Password,
			`-D`, cfg.Database.MySQL.Database,
		)
		fp, err = os.Open(`setup/data/schemas.mysql.sql`)
		if err != nil {
			panic(err)
		}
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
		now := datetime.MyLocal()
		tdb.Model(&models.Option{
			Name:  `db_ver`,
			Value: fmt.Sprint(dbVer),
		}).MustCreate()
		tdb.Model(&models.Category{
			Name:     `未分类`,
			Slug:     `uncategorized`,
			ParentID: 0,
			Path:     `/`,
		}).MustCreate()
		tdb.Model(&models.Post{
			Date:       now,
			Modified:   now,
			Title:      `你好，世界`,
			Content:    `你好，世界！这是您的第一篇文章。`,
			Type:       `post`,
			Category:   1,
			Status:     `public`,
			SourceType: `markdown`,
			Source:     `你好，世界！这是您的第一篇文章。`,
		}).MustCreate()
	})
}
