package migration

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/movsb/taoblog/service/models"
	setup_data "github.com/movsb/taoblog/setup/data"
	"github.com/movsb/taorm"
)

// Init ...
func Init(db *sql.DB, path string) {
	var err error

	fp, err := setup_data.Root.Open(`schemas.sqlite.sql`)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	all, err := io.ReadAll(fp)
	if err != nil {
		panic(err)
	}

	tdb := taorm.NewDB(db)

	result, err := tdb.Exec(string(all))
	if err != nil {
		panic(err)
	}
	_ = result

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
