package migration

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
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
// ID 总是 2；1 留给 SystemAdmin 了。
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
