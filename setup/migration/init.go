package migration

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	setup_data "github.com/movsb/taoblog/setup/data"
	"github.com/movsb/taorm"
)

// 初始化文章数据库。
func InitPosts(path string, createFirstPost bool) *sql.DB {
	db := createDatabase(path)
	testPosts(db, createFirstPost)
	return db
}

// 初始化文件数据库。
func InitFiles(path string) *sql.DB {
	db := createDatabase(path)
	testFiles(db)
	return db
}

// 如果路径为空，使用内存数据库。
func createDatabase(path string) *sql.DB {
	var db *sql.DB
	var err error

	v := url.Values{}
	v.Set(`cache`, `shared`)
	v.Set(`mode`, `rwc`)

	if path == `` {
		// 内存数据库
		// NOTE: 测试的时候同名路径会引用同一个内存数据库，
		// 所以需要取不同的路径名。
		path = fmt.Sprintf(`%s@%d`,
			`no-matter-what-path-used`,
			time.Now().UnixNano(),
		)
		v.Set(`mode`, `memory`)
	}

	u := url.URL{
		Scheme:   `file`,
		Opaque:   url.PathEscape(path),
		RawQuery: v.Encode(),
	}

	dsn := u.String()
	// log.Println(`数据库连接字符串：`, dsn)
	db, err = sql.Open(`sqlite3`, dsn)
	if err == nil {
		db.SetMaxOpenConns(1)
	}
	if err != nil {
		panic(err)
	}

	return db
}

// 初始化过程。会创建所有的数据表。
// 会自动创建管理员用户。
// TODO 移除用户创建。
func initPosts(db *sql.DB) {
	var err error

	fp, err := setup_data.Root.Open(`posts.sql`)
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

func initFiles(db *sql.DB) {
	var err error

	fp, err := setup_data.Root.Open(`files.sql`)
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
}

func testPosts(db *sql.DB, createFirstPost bool) {
	var count int
	row := db.QueryRow(`select count(1) from options`)
	if err := row.Scan(&count); err != nil {
		if se, ok := err.(sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				initPosts(db)

				tdb := taorm.NewDB(db)
				now := time.Now().Unix()

				if createFirstPost {
					tdb.MustTxCall(func(tx *taorm.DB) {
						tx.Model(&models.Post{
							UserID:     int32(auth.AdminID),
							Date:       int32(now),
							Modified:   int32(now),
							Title:      `你好，世界`,
							Type:       `post`,
							Category:   0,
							Status:     `public`,
							SourceType: `markdown`,
							Source:     `你好，世界！这是您的第一篇文章。`,

							// TODO 用配置时区。
							DateTimezone: ``,
							// TODO 用配置时区。
							ModifiedTimezone: ``,
						}).MustCreate()
					})
				}
				return
			}
		}
		panic(err)
	}
}

func testFiles(db *sql.DB) {
	var count int
	row := db.QueryRow(`select count(1) from files`)
	if err := row.Scan(&count); err != nil {
		if se, ok := err.(sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				initFiles(db)
				return
			}
		}
		panic(err)
	}
}
