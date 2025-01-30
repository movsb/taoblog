package migration

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	setup_data "github.com/movsb/taoblog/setup/data"
	"github.com/movsb/taorm"
)

// 初始化过程。会创建所有的数据表。
// 会自动创建管理员用户。
// TODO 移除用户创建。
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

	now := int32(time.Now().Unix())

	tdb.MustTxCall(func(tx *taorm.DB) {
		tx.Model(&models.Option{
			Name:  `db_ver`,
			Value: fmt.Sprint(MaxVersionNumber()),
		}).MustCreate()
	})

	tdb.MustTxCall(func(tx *taorm.DB) {
		var r [16]byte
		utils.Must1(rand.Read(r[:]))
		user := models.User{
			ID:        2,
			CreatedAt: int64(now),
			UpdatedAt: int64(now),
			Password:  fmt.Sprintf(`%x`, r),
		}
		tx.Model(&user).MustCreate()
		log.Println(`管理员密码：`, user.Password)
	})
}
