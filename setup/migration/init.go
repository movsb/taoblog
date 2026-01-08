package migration

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	setup_data "github.com/movsb/taoblog/setup/data"
	"github.com/movsb/taorm"

	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// 初始化文章数据库。
func InitPosts(path string, testCompat bool, createFirstPost bool) *sql.DB {
	db := createDatabase(path)
	testPosts(db, testCompat, createFirstPost)
	return db
}

// 初始化文件数据库。
//
// 注意：这里只是初始化文件实际内容存储的数据库，文件元数据存储在文章数据库中。
func InitFiles(path string) *sql.DB {
	db := createDatabase(path)
	testFiles(db)
	return db
}

// 初始化缓存数据库。
func InitCache(path string) *sql.DB {
	db := createDatabase(path)
	testCache(db)
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
	_initSQL(db, `posts.sql`)

	now := int32(time.Now().Unix())

	tdb := taorm.NewDB(db)

	tdb.MustTxCall(func(tx *taorm.DB) {
		tx.Model(&models.Option{
			Name:  `db_ver`,
			Value: fmt.Sprint(MaxVersionNumber()),
		}).MustCreate()
	})

	// 写入建站时间。
	tdb.MustTxCall(func(tx *taorm.DB) {
		tx.Model(&models.Option{
			Name:  `site.since`,
			Value: fmt.Sprint(now),
		}).MustCreate()
	})

	tdb.MustTxCall(func(tx *taorm.DB) {
		var r [16]byte
		utils.Must1(rand.Read(r[:]))
		user := models.User{
			ID:        2,
			CreatedAt: int64(now),
			UpdatedAt: int64(now),
			Nickname:  `管理员`,
			Password:  fmt.Sprintf(`%x`, r),
		}
		tx.Model(&user).MustCreate()
		log.Printf(`管理员用户名和密码：%d %s（仅首次运行出现）`, user.ID, user.Password)
	})
}

func initFiles(db *sql.DB) {
	_initSQL(db, `files.sql`)
}

func initCache(db *sql.DB) {
	_initSQL(db, `cache.sql`)
}

func _initSQL(db *sql.DB, file string) {
	var err error
	fp, err := setup_data.Root.Open(file)
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

func testPosts(db *sql.DB, testCompat bool, createFirstPost bool) {
	var ver string
	row := db.QueryRow(`select value from options where name='db_ver'`)
	if err := row.Scan(&ver); err != nil {
		if se, ok := err.(*sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				initPosts(db)

				tdb := taorm.NewDB(db)
				now := time.Now().Unix()

				if createFirstPost {
					tdb.MustTxCall(func(tx *taorm.DB) {
						tx.Model(&models.Post{
							UserID:     int32(user.AdminID),
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

	if testCompat {
		n, _ := strconv.Atoi(ver)
		if n != MaxVersionNumber() {
			log.Fatalln(`不兼容的数据库版本。`)
		}
	}
}

func testFiles(db *sql.DB) {
	var count int
	row := db.QueryRow(`select count(1) from files`)
	if err := row.Scan(&count); err != nil {
		if se, ok := err.(*sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				initFiles(db)
				return
			}
		}
		panic(err)
	}
}

func testCache(db *sql.DB) {
	var count int
	row := db.QueryRow(`select count(1) from cache`)
	if err := row.Scan(&count); err != nil {
		if se, ok := err.(*sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				initCache(db)
				return
			}
		}
		panic(err)
	}
}
