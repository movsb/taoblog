package migration

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/renderers/exif/exif_exports"
	"github.com/movsb/taoblog/service/modules/renderers/media_size"
	"github.com/movsb/taorm"
)

func v0(tx *sql.Tx) {

}

func v1(tx *sql.Tx) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	_, err = tx.Exec(`UPDATE posts SET source='' WHERE source IS NULL`)
	_, err = tx.Exec(`ALTER TABLE posts CHANGE source source LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL`)
}

func v2(tx *sql.Tx) {
	if _, err := tx.Exec(`DROP TABLE IF EXISTS shuoshuo;`); err != nil {
		panic(err)
	}
	if _, err := tx.Exec(`DROP TABLE IF EXISTS shuoshuo_comments;`); err != nil {
		panic(err)
	}
}

func v3(tx *sql.Tx) {
	type v3PostMetas struct {
		ID       int64
		Type     string
		Tid      int64
		Header   string
		Footer   string
		Keywords string
	}
	type v3Meta struct {
		ID    int64  `json:"id"`
		Metas string `json:"metas"`
	}
	var metas []*v3PostMetas
	if err := taorm.ScanRows(&metas, tx, `SELECT * FROM post_metas WHERE type='post' OR type='page'`); err != nil {
		panic(err)
	}
	for _, meta := range metas {
		var pm v3Meta
		query := `SELECT id,metas FROM posts WHERE id=?`
		if err := taorm.ScanRows(&pm, tx, query, meta.Tid); err != nil {
			panic(err)
		}
		var m map[string]any
		if pm.Metas == "" {
			pm.Metas = "{}"
		}
		if err := json.Unmarshal([]byte(pm.Metas), &m); err != nil {
			panic(err)
		}
		m["header"] = meta.Header
		m["footer"] = meta.Footer
		by, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}
		pm.Metas = string(by)
		_, err = tx.Exec(`UPDATE posts SET metas=? WHERE id=?`, pm.Metas, pm.ID)
		if err != nil {
			panic(err)
		}
	}
	_, err := tx.Exec(`DROP TABLE post_metas`)
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec(`UPDATE posts SET metas='{}' WHERE metas=''`)
	if err != nil {
		panic(err)
	}
}

func v4(tx *sql.Tx) {
	s := "CREATE UNIQUE INDEX `uix_post_id_and_tag_id` ON `post_tags` (`post_id`, `tag_id`)"
	if _, err := tx.Exec(s); err != nil {
		panic(err)
	}
}

func v5(tx *sql.Tx) {
	s := "CREATE UNIQUE INDEX `uix_name` ON `options` (`name`)"
	if _, err := tx.Exec(s); err != nil {
		panic(err)
	}
}

func v6(tx *sql.Tx) {
	ss := []string{
		"UPDATE posts SET date=DATE_ADD(date, INTERVAL 8 HOUR),modified=DATE_ADD(modified, INTERVAL 8 HOUR)",
		"UPDATE comments SET date=DATE_ADD(date, INTERVAL 8 HOUR)",
	}
	for _, s := range ss {
		if _, err := tx.Exec(s); err != nil {
			panic(err)
		}
	}
}

func v7(tx *sql.Tx) {
	var login string
	query := "SELECT value FROM options WHERE name=?"
	row := tx.QueryRow(query, "login")
	if err := row.Scan(&login); err != nil {
		panic(err)
	}
	query = "UPDATE options SET value=? WHERE name=?"
	if _, err := tx.Exec(query, login, "login"); err != nil {
		panic(err)
	}
}

func v8(tx *sql.Tx) {
	q := `ALTER TABLE comments CHANGE ancestor root INT(20) UNSIGNED NOT NULL`
	if _, err := tx.Exec(q); err != nil {
		panic(err)
	}
}

func v9(tx *sql.Tx) {
	q := `ALTER TABLE taxonomies CHANGE ancestor root INT(20) UNSIGNED NOT NULL`
	if _, err := tx.Exec(q); err != nil {
		panic(err)
	}
}

func v10(tx *sql.Tx) {
	queries := []string{
		"RENAME TABLE `taxonomies` TO `categories`",
		"ALTER TABLE `categories` ADD COLUMN `path` VARCHAR(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL",
	}
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			panic(err)
		}
	}

	type v10Category struct {
		ID     uint
		Name   string
		Slug   string
		Parent uint
		Root   uint
		Path   string
	}

	var setPaths func(cat *v10Category, path string)
	setPaths = func(cat *v10Category, path string) {
		query := `UPDATE categories SET path=? WHERE id=?`
		if _, err := tx.Exec(query, path, cat.ID); err != nil {
			panic(err)
		}
		var children []*v10Category
		query = `SELECT * FROM categories WHERE parent=?`
		taorm.MustScanRows(&children, tx, query, cat.ID)
		childPath := path + "/" + cat.Slug
		if path == "/" {
			childPath = childPath[1:]
		}
		for _, child := range children {
			setPaths(child, childPath)
		}
	}

	var topLevels []*v10Category
	q := `SELECT * FROM categories WHERE parent = 0`
	taorm.MustScanRows(&topLevels, tx, q)
	for _, cat := range topLevels {
		setPaths(cat, "/")
	}

	queries = []string{
		"ALTER TABLE `categories` CHANGE `parent` `parent_id` INT(10) UNSIGNED NOT NULL",
		"ALTER TABLE `categories` DROP `root`",
		"DELETE FROM `categories` WHERE `path` = ''",
		"CREATE INDEX `uix_path_slug` ON `categories` (`path`,`slug`)",
		"ALTER TABLE `posts` CHANGE `taxonomy` `category` INT(10) UNSIGNED NOT NULL DEFAULT 1",
	}
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			panic(err)
		}
	}
}

func v11(tx *sql.Tx) {
	tx.Exec(`DELETE FROM options WHERE name = ?`, `home`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `blog_name`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `blog_desc`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `keywords`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `login`)
}

func v12(tx *sql.Tx) {
	tx.Exec("ALTER TABLE comments ADD COLUMN `source_type` varchar(16) NOT NULL AFTER `date`")
	tx.Exec("ALTER TABLE comments ADD COLUMN `source` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL AFTER `source_type`")
	tx.Exec("UPDATE comments SET source_type='plain'")
}

func v13(tx *sql.Tx) {
	tx.Exec(`DELETE FROM options WHERE name = ?`, `email`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `not_allowed_emails`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `not_allowed_authors`)
}

func v14(tx *sql.Tx) {
	tx.Exec(`UPDATE posts SET comments = 0`)
	tx.Exec(`UPDATE posts INNER JOIN (SELECT post_id,count(id) AS comments FROM comments GROUP BY post_id) AS counts ON posts.id = counts.post_id SET posts.comments = counts.comments`)
}

func v15(tx *sql.Tx) {
	if _, err := tx.Exec(`DELETE FROM categories WHERE slug=? AND parent_id=?`, `uncategorized`, 0); err != nil {
		panic(err)
	}
	if _, err := tx.Exec(`UPDATE posts SET category=? WHERE category=?`, 0, 1); err != nil {
		panic(err)
	}
}

func v16(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE comments ADD COLUMN `date_int` INT NOT NULL AFTER `date`")
	mustExec(tx, "UPDATE comments SET `date_int` = UNIX_TIMESTAMP(`date`)")
	mustExec(tx, "ALTER TABLE comments DROP COLUMN `date`")
	mustExec(tx, "ALTER TABLE comments CHANGE `date_int` `date` INT NOT NULL")

	mustExec(tx, "ALTER TABLE posts ADD COLUMN `date_int` INT NOT NULL AFTER `date`")
	mustExec(tx, "UPDATE posts SET `date_int` = UNIX_TIMESTAMP(`date`)")
	mustExec(tx, "ALTER TABLE posts DROP COLUMN `date`")
	mustExec(tx, "ALTER TABLE posts CHANGE `date_int` `date` INT NOT NULL")

	mustExec(tx, "ALTER TABLE posts ADD COLUMN `modified_int` INT NOT NULL AFTER `modified`")
	mustExec(tx, "UPDATE posts SET `modified_int` = UNIX_TIMESTAMP(`modified`)")
	mustExec(tx, "ALTER TABLE posts DROP COLUMN `modified`")
	mustExec(tx, "ALTER TABLE posts CHANGE `modified_int` `modified` INT NOT NULL")
}

// SQLite3 only

func v17(tx *sql.Tx) {
	mustExec(tx, `
	CREATE TABLE options2 (
		id INTEGER  PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(64)  NOT NULL UNIQUE COLLATE NOCASE,
		value TEXT  NOT NULL
	)
	`)
	mustExec(tx, `INSERT INTO options2 SELECT * FROM options`)
	mustExec(tx, `DROP TABLE options`)
	mustExec(tx, `ALTER TABLE options2 RENAME TO options`)

	mustExec(tx, `
	CREATE TABLE tags2 (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE COLLATE NOCASE,
		alias INTEGER NOT NULL
	)
	`)
	mustExec(tx, `INSERT INTO tags2 SELECT * FROM tags`)
	mustExec(tx, `DROP TABLE tags`)
	mustExec(tx, `ALTER TABLE tags2 RENAME TO tags`)
}

func v18(tx *sql.Tx) {
	mustExec(tx, `CREATE TABLE IF NOT EXISTS pingbacks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at INTEGER  NOT NULL,
		post_id INTEGER  NOT NULL,
		title TEXT NOT NULL,
		source_url TEXT NOT NULL,
		UNIQUE (post_id, source_url)
	)`)
}

func v19(tx *sql.Tx) {
	mustExec(tx, `CREATE TABLE IF NOT EXISTS redirects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at INTEGER  NOT NULL,
		source_path TEXT NOT NULL,
		target_path TEXT NOT NULL,
		status_code INTEGER NOT NULL,
		UNIQUE (source_path)
	)`)
}

func v20(tx *sql.Tx) {
	mustExec(tx, "UPDATE posts SET source=content WHERE source='' AND source_type='html'")
	mustExec(tx, "ALTER TABLE posts DROP COLUMN `content`")
}

func v21(tx *sql.Tx) {
	mustExec(tx, `DROP TABLE pingbacks`)
}

type MapStr2Str map[string]string

func (m MapStr2Str) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MapStr2Str) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	}
	return errors.New(`unsupported type`)
}

func v22(tx *sql.Tx) {
	type v22PostOld struct {
		ID    int64
		Metas MapStr2Str
	}

	// 暂时有 Bugs，必须用指针。
	var oldPosts []*v22PostOld
	taorm.MustScanRows(&oldPosts, tx, `select id,metas from posts where  metas not like 'null' and metas != '{}'`)

	type v22PostNew struct {
		ID    int64
		Metas models.PostMeta
	}

	var newPosts []v22PostNew
	for _, old := range oldPosts {
		new := v22PostNew{
			ID:    old.ID,
			Metas: models.PostMeta{},
		}
		m := &new.Metas

		for key, val := range old.Metas {
			switch key {
			case `header`:
				m.Header = val
			case `footer`:
				m.Footer = val
			case `outdated`:
				m.Outdated = val == `true` || val == `1`
			case `wide`:
				m.Wide = val == `true` || val == `1`
			case `weixin`:
				m.Weixin = val
			default:
				panic(fmt.Sprintf("未知元数据：%s: %v", key, val))
			}
		}

		newPosts = append(newPosts, new)
	}

	for _, new := range newPosts {
		mustExec(tx, `update posts set metas=? where id=?`, new.Metas, new.ID)
	}
}

func v23(tx *sql.Tx) {
	mustExec(tx, "CREATE INDEX `post_id` on `comments` (`post_id`)")
}

// 由于评论更新并不会更新文章的修改时间，但页面应该重新渲染。
// 文章的修改时间（特指用于处理 304 的那个时间），应该用两者中的较大者。
// 所以这个功能是为了记录最后的评论时间。
// ~~但是这个时间我觉得并没有必要存到数据库里面。~~
// 更新：有必要。因为历史评论可能被编辑，被编辑的评论会影响文章的“修改时间”，但是不能更新到评论本身，
// 否则评论列表就会变化。因为目前评论只有创建时间，没有修改时间。评论列表是按创建时间排序的。
// 下面这条语句本来是想执行的，但是报错。不执行问题也不大。
/*
	mustExec(tx, `
UPDATE
	posts
SET
	last_comment_time = dates.max_date
INNER JOIN
	comments
ON
	posts.id = comments.post_id
INNER JOIN
	(
		SELECT
			max(date) AS max_date
		FROM
			comments
		GROUP BY
			post_id
	) AS dates
ON comments.id = dates.id AND comments.date = dates.max_date
`)
*/
func v24(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE posts ADD COLUMN `last_commented_at` INTEGER NOT NULL DEFAULT 0")
}

func v25(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE comments ADD COLUMN `modified` INTEGER NOT NULL DEFAULT 0")
	mustExec(tx, "UPDATE comments set modified = date")
}

func v26(tx *sql.Tx) {
	mustExec(tx, "UPDATE comments SET source=content WHERE source_type='plain' AND source=''")
	mustExec(tx, "ALTER TABLE comments DROP COLUMN content")
}

func v27(tx *sql.Tx) {
	mustExec(tx, "UPDATE posts SET title='' WHERE title='无标题' OR title='Untitled'")
}

func v28(tx *sql.Tx) {
	mustExec(tx, `DROP TABLE categories`)
	mustExec(tx, `DROP TABLE redirects`)
}

func v29(tx *sql.Tx) {
	type v29Comment struct {
		ID     int
		Root   int
		Parent int
	}

	var all []*v29Comment
	taorm.MustScanRows(&all, tx, `SELECT id,root,parent FROM comments`)

	slices.SortFunc(all, func(a, b *v29Comment) int {
		return a.ID - b.ID
	})

	ids := map[int]struct{}{}
	for _, c := range all {
		ids[c.ID] = struct{}{}
	}

	var toDelete []int

	for _, c := range all {
		if c.Parent == 0 {
			continue
		}
		if _, ok := ids[c.Parent]; !ok {
			delete(ids, c.ID)
			toDelete = append(toDelete, c.ID)
		}
	}

	for _, id := range toDelete {
		log.Println(`将删除评论：`, id)
		mustExec(tx, `delete from comments where id=?`, id)
	}
}

func v30(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE posts ADD COLUMN `date_timezone` TEXT NOT NULL DEFAULT ''")
	mustExec(tx, "ALTER TABLE posts ADD COLUMN `modified_timezone` TEXT NOT NULL DEFAULT ''")
	mustExec(tx, "ALTER TABLE comments ADD COLUMN `date_timezone` TEXT NOT NULL DEFAULT ''")
	mustExec(tx, "ALTER TABLE comments ADD COLUMN `modified_timezone` TEXT NOT NULL DEFAULT ''")
}

func v31(tx *sql.Tx) {
	mustExec(tx, `DELETE FROM options WHERE name = 'vps.hostdare'`)
	mustExec(tx, `DELETE FROM options WHERE name = 'vps.current'`)
}

func v32(tx *sql.Tx) {
	mustExec(tx, "CREATE TABLE IF NOT EXISTS logs (`id` INTEGER PRIMARY KEY AUTOINCREMENT,`time` INTEGER NOT NULL,`type` TEXT NOT NULL COLLATE NOCASE,`sub_type` TEXT NOT NULL COLLATE NOCASE,`version` INTEGER NOT NULL,`data` TEXT NOT NULL)")
}

func v33(tx *sql.Tx) {
	mustExec(tx, "CREATE TABLE IF NOT EXISTS users (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `created_at` INTEGER NOT NULL, `updated_at` INTEGER NOT NULL, `password` TEXT NOT NULL)")
}

var AdminID = 2

// 密码是随机生成的，会同时使用配置文件里面的 key 和此 password，
// 待到后续移除 key 后，就仅使用此 password 了。
// ID 总是 2；1 留给 System 了。
func v34(tx *sql.Tx) {
	var r [16]byte
	utils.Must1(rand.Read(r[:]))
	password := fmt.Sprintf(`%x`, r)

	// 暂时使用当前时间，其实应该是建站时间。
	now := time.Now().Unix()

	mustExec(tx, `INSERT INTO users (id,created_at,updated_at,password) VALUES (?,?,?,?)`, AdminID, now, now, password)
}

func v35(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE users ADD COLUMN `credentials` TEXT NOT NULL DEFAULT ''")
	var credentials string
	if err := tx.QueryRow(`SELECT value FROM options WHERE name=?`, `admin_webauthn_credentials`).Scan(&credentials); err == nil {
		mustExec(tx, `UPDATE users SET credentials=? WHERE id=?`, credentials, AdminID)
	}
	mustExec(tx, `DELETE FROM options WHERE name=?`, `admin_webauthn_credentials`)
}

func v36(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE users ADD COLUMN `google_user_id` TEXT NOT NULL DEFAULT ''")
	mustExec(tx, "ALTER TABLE users ADD COLUMN `github_user_id` INTEGER NOT NULL DEFAULT 0")
}

func v37(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE posts ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0")
	mustExec(tx, "UPDATE posts SET user_id=?", AdminID)
}

func v38(tx *sql.Tx) {
	mustExec(tx, "CREATE TABLE IF NOT EXISTS acl ( `id` INTEGER PRIMARY KEY AUTOINCREMENT, `created_at` INTEGER NOT NULL, `post_id` INTEGER NOT NULL, `user_id` INTEGER NOT NULL, `permission` TEXT NOT NULL)")
}

func v39(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE users ADD COLUMN nickname TEXT NOT NULL DEFAULT ''")
	mustExec(tx, "ALTER TABLE users ADD COLUMN email TEXT NOT NULL DEFAULT ''")
}

func v40(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE comments ADD COLUMN user_id INTEGER NOT NULL DEFAULT 0")
}

func v41(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE users ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0")
}

func v42(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE users ADD COLUMN avatar TEXT NOT NULL DEFAULT ''")
}

func v43(tx *sql.Tx) {
	mustExec(tx, "DELETE FROM options WHERE name='exif:cache'")
}
func v44(tx *sql.Tx) {
	mustExec(tx, "DELETE FROM options WHERE name='friends:cache'")
}

// 有一个给 files 增加 meta 的操作没有记录。
// 涉及到多数据库，暂时不支持。
// ALTER TABLE files ADD COLUMN meta BLOB NOT NULL DEFAULT '{}' AFTER `size`;

func v45(posts, files, cache *taorm.DB) {
	files.MustExec("ALTER TABLE files ADD COLUMN digest TEXT NOT NULL DEFAULT ''")
	var ids []struct{ ID int }
	files.Raw(`SELECT id FROM files`).MustFind(&ids)
	for _, id := range ids {
		var data string // []byte
		files.Raw(`SELECT data FROM files WHERE id=?`, id.ID).MustFind(&data)
		d := md5.New()
		fmt.Fprint(d, data)
		s := d.Sum(nil)
		x := fmt.Sprintf(`%x`, s)
		files.MustExec(`UPDATE files SET digest=? WHERE id=?`, x, id.ID)
		log.Println(`文件摘要：`, id.ID, x)
	}
}

func v46(posts, files, cache *taorm.DB) {
	var list []struct{ ID int }
	files.Raw(`SELECT id FROM files WHERE digest=''`).MustFind(&list)
	for _, file := range list {
		var data string // []byte
		files.Raw(`SELECT data FROM files WHERE id=?`, file.ID).MustFind(&data)
		d := md5.New()
		fmt.Fprint(d, data)
		s := d.Sum(nil)
		x := fmt.Sprintf(`%x`, s)
		files.MustExec(`UPDATE files SET digest=? WHERE id=?`, x, file.ID)
		log.Println(`文件摘要：`, file.ID, x)
	}
}

func v47(posts, files, cache *taorm.DB) {
	posts.MustExec("CREATE INDEX `idx_modified` ON `posts` (`modified`)")
}

func v48(posts, files, cache *taorm.DB) {
	posts.MustExec("ALTER TABLE posts ADD COLUMN citations TEXT NOT NULL DEFAULT '{}'")
}

func v49(posts, files, cache *taorm.DB) {
	log.Println(`删除索引并创建备份表……`)
	files.MustExec(`DROP INDEX post_id__path`)
	files.MustExec("CREATE TABLE `files_copy` (`id` INTEGER  PRIMARY KEY AUTOINCREMENT,`created_at` INTEGER NOT NULL,`updated_at` INTEGER NOT NULL,`post_id` INTEGER NOT NULL,`path` TEXT NOT NULL,`mode` INTEGER NOT NULL,`mod_time` INTEGER  NOT NULL,`size` INTEGER  NOT NULL,`meta` BLOB NOT NULL,`digest` TEXT NOT NULL,`data` BLOB NOT NULL)")
	log.Println(`正在拷贝旧数据到新表……`)
	files.MustExec(`INSERT INTO files_copy (id,created_at,updated_at,post_id,path,mode,mod_time,size,meta,digest,data) SELECT id,created_at,updated_at,post_id,path,mode,mod_time,size,meta,digest,data FROM files`)
	log.Println(`正在删除旧表……`)
	files.MustExec(`DROP TABLE files`)
	log.Println(`重命名新表为旧表……`)
	files.MustExec(`ALTER TABLE files_copy RENAME TO files`)
	log.Println(`创建新的索引……`)
	files.MustExec("CREATE UNIQUE INDEX `post_id__path` ON `files` (`post_id`,`path`)")
}

type v50Option struct {
	ID    int
	Name  string
	Value string
}

func (v50Option) TableName() string {
	return `options`
}

// site.sync.r2
// {"Enabled":false,"AccountID":"","AccessKeyID":"","AccessKeySecret":"","BucketName":""}
func v50(posts, files, cache *taorm.DB) {
	type v50R2Config struct {
		Enabled         bool
		AccountID       string
		AccessKeyID     string
		AccessKeySecret string
		BucketName      string
	}
	type v50OSSConfig struct {
		Enabled         bool
		Endpoint        string
		Region          string
		AccessKeyID     string
		AccessKeySecret string
		BucketName      string
	}
	var opt v50Option
	if err := posts.From(models.Option{}).Where(`name='site.sync.r2'`).Find(&opt); err != nil {
		if taorm.IsNotFoundError(err) {
			return
		}
		panic(err)
	}
	var r2 v50R2Config
	utils.Must(json.Unmarshal([]byte(opt.Value), &r2))
	oss := v50OSSConfig{
		Enabled:         r2.Enabled,
		Endpoint:        fmt.Sprintf(`https://%s.r2.cloudflarestorage.com`, r2.AccountID),
		Region:          `auto`,
		AccessKeyID:     r2.AccessKeyID,
		AccessKeySecret: r2.AccessKeySecret,
		BucketName:      r2.BucketName,
	}
	by := utils.Must1(json.Marshal(oss))
	posts.Model(&opt).MustUpdateMap(taorm.M{
		`value`: string(by),
	})
}

func v51(posts, files, cache *taorm.DB) {
	posts.MustExec(`DELETE FROM options WHERE name='avatar:cache'`)
}

func v52(posts, files, cache *taorm.DB) {
	var list []struct{ ID int }
	files.Raw(`SELECT id FROM files`).MustFind(&list)
	for _, file := range list {
		var data string // []byte
		files.Raw(`SELECT data FROM files WHERE id=?`, file.ID).MustFind(&data)
		d := md5.New()
		d.Write([]byte(data))
		s := d.Sum(nil)
		x := fmt.Sprintf(`%x`, s)
		files.MustExec(`UPDATE files SET digest=? WHERE id=?`, x, file.ID)
		log.Println(`文件摘要：`, file.ID, x)
	}
}

type v53File struct {
	Data string
	Meta models.FileMeta
}

func (v53File) TableName() string {
	return `files`
}

func v53(posts, files, cache *taorm.DB) {
	var list []struct{ ID int }
	files.Raw(`SELECT id FROM files`).MustFind(&list)
	for _, file := range list {
		var model v53File
		files.Raw(`SELECT meta,data FROM files WHERE id=?`, file.ID).MustFind(&model)
		models.Encrypt(&model.Meta.Encryption, []byte(model.Data))
		files.MustExec(`UPDATE files SET meta=? WHERE id=?`, model.Meta, file.ID)
		log.Println(`文件摘要：`, file.ID, model.Meta.Encryption)
	}
}

func v54(posts, files, cache *taorm.DB) {
	posts.MustExec("ALTER TABLE users ADD COLUMN otp_secret TEXT NOT NULL DEFAULT ''")
}

func v55(posts, files, cache *taorm.DB) {
	posts.MustExec("ALTER TABLE users ADD COLUMN chanify_token TEXT NOT NULL DEFAULT ''")
}

func v56(posts, files, cache *taorm.DB) {
	var list []*models.File
	files.Select(`id,path,meta`).MustFind(&list)
	for _, f := range list {
		if f.Meta.Width != 0 {
			continue
		}
		var d string
		var w, h int
		files.Raw(`select data from files where id=?`, f.ID).MustFind(&d)
		m, err := media_size.All(strings.NewReader(d))
		if err != nil {
			m2, err2 := exif_exports.Extract(io.NopCloser(strings.NewReader(d)))
			if err2 != nil {
				log.Println(f.ID, f.Path, err)
				continue
			}
			if _, err := fmt.Sscanf(m2.ImageSize, `%dx%d`, &w, &h); err != nil {
				log.Println(f.ID, f.Path, err)
				continue
			}
		} else {
			w = m.Width
			h = m.Height
		}
		f.Meta.Width = w
		f.Meta.Height = h
		files.MustExec(`update files set meta=? where id=?`, f.Meta, f.ID)
	}
}

func v57(posts, files, cache *taorm.DB) {
	posts.MustExec(`ALTER TABLE users RENAME COLUMN chanify_token TO bark_token`)
	posts.MustExec(`DELETE FROM options WHERE name=?`, `notify.chanify`)
}

func v58(posts, files, cache *taorm.DB) {
	posts.MustExec(`CREATE TABLE categories (
		id INTEGER  PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT  NOT NULL COLLATE NOCASE
	)`)
	posts.MustExec("CREATE UNIQUE INDEX `uix_cat_user_id__name` ON `categories` (`user_id`,`name`)")
}

// 分类被重新加回来了，以前的分类表已被删除，所以全部重置。
func v59(posts, files, cache *taorm.DB) {
	t := time.Date(2025, 5, 30, 0, 0, 0, 0, time.Local)
	posts.MustExec("UPDATE posts SET category=0 WHERE date < ?", t.Unix())
}

type v60FileInfo struct {
	ID        int
	CreatedAt int64
	UpdatedAt int64
	PostID    int
	Path      string
	Mode      uint32
	ModTime   int64
	Size      uint32
	Digest    string
	Meta      []byte
}

func (v60FileInfo) TableName() string {
	return `files`
}

type v60FileOld struct {
	v60FileInfo
	Data []byte
}

func (v60FileOld) TableName() string {
	return `files`
}

type v60FileData struct {
	ID     int
	PostID int
	Digest string
	Data   []byte
}

func (v60FileData) TableName() string {
	return `files_new`
}

// 几个步骤：
//  1. 创建新表 files
//  2. 从旧表 files 复制基本信息到新表 files
//  3. 更新旧表 files：只保留必要字段
func v60(posts, files, cache *taorm.DB) {
	log.Println(`创建新表 files...`)
	posts.MustExec("CREATE TABLE IF NOT EXISTS `files` (`id` INTEGER  PRIMARY KEY AUTOINCREMENT,`created_at` INTEGER NOT NULL,`updated_at` INTEGER NOT NULL,`post_id` INTEGER NOT NULL,`path` TEXT NOT NULL,`mode` INTEGER NOT NULL,`mod_time` INTEGER  NOT NULL,`size` INTEGER  NOT NULL,`digest` TEXT NOT NULL,`meta` BLOB NOT NULL)")
	posts.MustExec("CREATE UNIQUE INDEX `post_id__path` ON `files` (`post_id`,`path`)")

	log.Println(`创建新表 files_new...`)
	files.MustExec("CREATE TABLE IF NOT EXISTS `files_new` (`id` INTEGER  PRIMARY KEY AUTOINCREMENT,`post_id` INTEGER NOT NULL,`digest` TEXT NOT NULL,`data` BLOB NOT NULL)")
	files.MustExec("CREATE UNIQUE INDEX `post_id__digest` ON `files_new` (`post_id`,`digest`)")

	log.Println(`从旧表 files 复制基本信息到新表 files...`)
	var oldFileInfos []*v60FileOld
	files.Select(`id`).MustFind(&oldFileInfos)

	for _, _fi := range oldFileInfos {
		var file v60FileOld
		files.Where(`id=?`, _fi.ID).MustFind(&file)

		log.Println(`复制文件信息：`, file.PostID, file.Path)

		posts.Model(&file.v60FileInfo).MustCreate()

		files.Model(&v60FileData{
			PostID: file.PostID,
			Digest: file.Digest,
			Data:   file.Data,
		}).MustCreate()
	}

	log.Println(`完成复制，更新旧表 files...`)

	files.MustExec("DROP TABLE files")
	files.MustExec("ALTER TABLE files_new RENAME TO files")
	// 索引自动跟表走，无需更名。
}

func v61(posts, files, cache *taorm.DB) {
	posts.MustExec(`UPDATE posts SET status='private' WHERE status='draft'`)
}

type v62Post struct {
	ID    int
	Metas string
}

func (v62Post) TableName() string {
	return `posts`
}

// "origin:omitempty":null
func v62(posts, files, cache *taorm.DB) {
	var ps []*v62Post
	posts.From(models.Post{}).Select(`id,metas`).MustFind(&ps)

	r := strings.NewReplacer(
		`{"origin:omitempty":null}`, `{}`, // 只有它
		`{"origin:omitempty":null,`, `{`, // 在左边
		`,"origin:omitempty":null}`, `}`, // 在右边
		`,"origin:omitempty":null,`, `,`, // 在中间
		`"origin:omitempty":{`, `"origin":{`, // 有值的。
	)

	for _, p := range ps {
		p.Metas = r.Replace(p.Metas)
	}

	for _, p := range ps {
		posts.Model(p).MustUpdateMap(taorm.M{
			`metas`: p.Metas,
		})
	}
}

func v63(posts, files, cache *taorm.DB) {
	posts.MustExec(`ALTER TABLE files ADD COLUMN parent_id INTEGER NOT NULL DEFAULT 0`)
}

func v64(posts, files, cache *taorm.DB) {
	posts.MustExec(`ALTER TABLE files DROP COLUMN mode`)
}

func v65(posts, files, cache *taorm.DB) {
	var s struct{ Value int32 }
	if err := posts.Raw(`SELECT value FROM options WHERE name='site.since'`).Find(&s); err != nil {
		if !taorm.IsNotFoundError(err) {
			panic(err)
		}
		posts.MustExec(`INSERT INTO options (name,value) VALUES ('site.since','1419350400')`)
		return
	}
}

func v66(posts, files, cache *taorm.DB) {
	var s struct{ Value string }
	if err := posts.Raw(`SELECT value FROM options WHERE name='favicon'`).Find(&s); err != nil {
		if !taorm.IsNotFoundError(err) {
			panic(err)
		}
		return
	}

	raw, _ := base64.RawURLEncoding.DecodeString(s.Value)
	u := utils.CreateDataURL(raw)
	posts.MustExec(`UPDATE options SET value=? WHERE name='favicon'`, u.String())
}

func v67(posts, files, cache *taorm.DB) {
	posts.MustExec(`DELETE FROM options WHERE name='blurhash:last'`)
}

func v68(posts, files, cache *taorm.DB) {
	posts.MustExec(`DELETE FROM options WHERE name='others.geo.baidu'`)
}
