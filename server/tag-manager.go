package main

import (
	"fmt"
	"log"
	"strings"
)

// Tag is a tag.
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Alias int64  `json:"alias"`
}

// TagManager manages tags.
type TagManager struct {
}

// NewTagManager news
func NewTagManager() *TagManager {
	return &TagManager{}
}

func (tm *TagManager) addTag(tx Querier, name string, alias uint) int64 {
	sql := `INSERT INTO tags (name,alias) values (?,?)`
	ret, err := tx.Exec(sql, name, alias)
	if err != nil {
		panic(err)
	}

	id, err := ret.LastInsertId()
	return id
}

func (tm *TagManager) searchTag(tx Querier, tag string) (tags []Tag) {
	sql := `SELECT * FROM tags WHERE name LIKE ?`
	rows, err := tx.Query(sql, "%"+tag+"%")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var tag Tag
		if err = rows.Scan(&tag.ID, &tag.Name, &tag.Alias); err != nil {
			panic(err)
		}
		tags = append(tags, tag)
	}

	return
}

func (tm *TagManager) getTagID(tx Querier, name string) int64 {
	sql := `SELECT id FROM tags WHERE name=? LIMIT 1`
	row := tx.QueryRow(sql, name)

	var id int64
	if err := row.Scan(&id); err != nil {
		log.Printf("标签名不存在：%s\n", name)
		id = 0
	}

	return id
}

func (tm *TagManager) hasTagName(tx Querier, name string) bool {
	return tm.getTagID(tx, name) > 0
}

func (tm *TagManager) getTagNames(tx Querier, pid int64) (names []string) {
	sql := `SELECT tags.name FROM post_tags,tags WHERE post_tags.post_id=` + fmt.Sprint(pid) + ` AND post_tags.tag_id=tags.id`
	rows, err := tx.Query(sql)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	names = make([]string, 0)

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			panic(err)
		}
		names = append(names, name)
	}

	return
}

func (tm *TagManager) getTagIDs(tx Querier, pid int64, alias bool) (ids []int64) {
	sql := `SELECT tag_id FROM post_tags WHERE post_id=` + fmt.Sprint(pid)

	rows, err := tx.Query(sql)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}

	if alias {
		ids = tm.getAliasTagsAll(tx, ids)
	}

	return
}

func (tm *TagManager) getAliasTagsAll(tx Querier, ids []int64) []int64 {
	sids := joinInts(ids, ",")

	sql1 := `SELECT alias FROM tags WHERE id in (?)`
	sql2 := `SELECT id FROM tags WHERE alias in (?)`

	rows, err := tx.Query(sql1, sids)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var alias int64
		if err = rows.Scan(&alias); err != nil {
			panic(err)
		}

		if alias > 0 {
			ids = append(ids, alias)
		}
	}

	rows.Close()

	rows, err = tx.Query(sql2, sids)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			panic(err)
		}

		ids = append(ids, id)
	}

	rows.Close()

	return ids
}

func (tm *TagManager) addObjectTag(tx Querier, pid int64, tid int64) int64 {
	sql := `INSERT INTO post_tags (post_id,tag_id) VALUES (?,?)`
	ret, err := tx.Exec(sql, pid, tid)
	if err != nil {
		panic(err)
	}

	id, err := ret.LastInsertId()

	return id
}

func (tm *TagManager) removeObjectTag(tx Querier, pid, tid int64) {
	sql := `DELETE FROM post_tags WHERE post_id=? AND tag_id=? LIMIT 1`
	ret, err := tx.Exec(sql, pid, tid)
	if err != nil {
		panic(err)
	}
	_ = ret
}

// UpdateObjectTags updates
func (tm *TagManager) UpdateObjectTags(tx Querier, pid int64, tagstr string) {
	// seperators are "," "，" ";" "；"
	tagstr = strings.Replace(tagstr, "，", ",", -1)
	tagstr = strings.Replace(tagstr, "；", ",", -1)
	tagstr = strings.Replace(tagstr, ";", ",", -1)

	newTags := strings.Split(tagstr, ",")
	oldTags := tm.getTagNames(tx, pid)

	var (
		toBeDeled []string
		toBeAdded []string
	)

	for _, t := range oldTags {
		if !strInSlice(newTags, t) {
			toBeDeled = append(toBeDeled, t)
		}
	}

	for _, t := range newTags {
		t = strings.TrimSpace(t)
		if t != "" && !strInSlice(oldTags, t) {
			toBeAdded = append(toBeAdded, t)
		}
	}

	for _, t := range toBeDeled {
		tid := tm.getTagID(tx, t)
		tm.removeObjectTag(tx, pid, tid)
	}

	for _, t := range toBeAdded {
		var tid int64
		if !tm.hasTagName(tx, t) {
			tid = tm.addTag(tx, t, 0)
		} else {
			tid = tm.getTagID(tx, t)
		}
		tm.addObjectTag(tx, pid, tid)
	}
}

// List lists all tags.
func (tm *TagManager) List(tx Querier) ([]*Tag, error) {
	rows, err := tx.Query("SELECT * FROM tags")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tags := make([]*Tag, 0)

	for rows.Next() {
		var tag Tag
		if err = rows.Scan(&tag.ID, &tag.Name, &tag.Alias); err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}

	return tags, rows.Err()
}
