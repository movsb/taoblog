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

// GetObjectTagNames gets all tag names of an object.
func (tm *TagManager) GetObjectTagNames(tx Querier, oid int64) ([]string, error) {
	q := make(map[string]interface{})
	q["select"] = "tags.name"
	q["from"] = "post_tags,tags"
	q["where"] = []string{
		"post_tags.post_id=" + fmt.Sprint(oid),
		"post_tags.tag_id=tags.id",
	}
	query := BuildQueryString(q)

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	names := make([]string, 0)

	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, rows.Err()
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
	oldTags, err := tm.GetObjectTagNames(tx, pid)
	if err != nil {
		return // TODO
	}

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
			tag, _ := tm.GetRootTag(tx, tm.getTagID(tx, t))
			tid = tag.ID
		}
		tm.addObjectTag(tx, pid, tid)
	}
}

// ListTags lists all tags.
func (tm *TagManager) ListTags(tx Querier) ([]*Tag, error) {
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

// GetTagByID gets a tag ID.
func (tm *TagManager) GetTagByID(tx Querier, id int64) (*Tag, error) {
	query := "SELECT id,name,alias FROM tags WHERE id=?"
	row := tx.QueryRow(query, id)
	var tag Tag
	if err := row.Scan(&tag.ID, &tag.Name, &tag.Alias); err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetRootTag gets the root tag of an alias-ed tag.
func (tm *TagManager) GetRootTag(tx Querier, id int64) (tag *Tag, err error) {
	for {
		tag, err = tm.GetTagByID(tx, id)
		if err != nil {
			return
		}
		if tag.Alias == 0 {
			return
		}
		id = tag.Alias
	}
}
