package service

import (
	"context"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm/taorm"
)

func (s *Service) tags() *taorm.Stmt {
	return s.tdb.Model(models.Tag{})
}

// GetTagByName gets a tag by Name.
func (s *Service) GetTagByName(name string) *models.Tag {
	var tag models.Tag
	err := s.tags().Where("name=?", name).Find(&tag)
	if err != nil {
		panic(&TagNotFoundError{})
	}
	return &tag
}

func (s *Service) ListTagsWithCount() []*models.TagWithCount {
	var tags []*models.TagWithCount
	s.tdb.Raw(`
		SELECT tags.name AS name, count(tags.id) AS count
		FROM tags INNER JOIN post_tags ON tags.id = post_tags.tag_id
		GROUP BY tags.id
		ORDER BY count desc
	`).MustFind(&tags)
	return tags
}

func (s *Service) getObjectTagIDs(postID int64, alias bool) (ids []int64) {
	sql := `SELECT tag_id FROM post_tags WHERE post_id=?`
	rows, err := s.tdb.Query(sql, postID)
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
		ids = s.getAliasTagsAll(ids)
	}

	return
}

// GetObjectTagNames ...
func (s *Service) GetObjectTagNames(postID int64) []string {
	query := `select tags.name from post_tags,tags where post_tags.post_id=? and post_tags.tag_id=tags.id`
	args := []interface{}{postID}
	rows, err := s.tdb.Query(query, args...)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	names := make([]string, 0)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			panic(err)
		}
		names = append(names, name)
	}
	return names
}

func (s *Service) getAliasTagsAll(ids []int64) []int64 {
	sids := utils.JoinInts(ids, ",")
	if sids == "" {
		return ids
	}

	sql1 := `SELECT alias FROM tags WHERE id in (?)`
	sql2 := `SELECT id FROM tags WHERE alias in (?)`

	rows, err := s.tdb.Query(sql1, sids)
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

	rows, err = s.tdb.Query(sql2, sids)
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

// UpdateObjectTags ...
func (s *Service) UpdateObjectTags(pid int64, tags []string) {
	newTags := tags
	oldTags := s.GetObjectTagNames(pid)

	var (
		toBeDeled []string
		toBeAdded []string
	)

	for _, t := range oldTags {
		if !utils.StrInSlice(newTags, t) {
			toBeDeled = append(toBeDeled, t)
		}
	}

	for _, t := range newTags {
		t = strings.TrimSpace(t)
		if t != "" && !utils.StrInSlice(oldTags, t) {
			toBeAdded = append(toBeAdded, t)
		}
	}

	for _, t := range toBeDeled {
		s.removeObjectTag(pid, t)
	}

	for _, t := range toBeAdded {
		var tid int64
		if !s.hasTagName(t) {
			tid = s.addTag(t)
		} else {
			tag := s.getRootTag(t)
			tid = tag.ID
		}
		s.addObjectTag(pid, tid)
	}
}

func (s *Service) removeObjectTag(pid int64, tagName string) {
	tagObj := s.GetTagByName(tagName)
	s.tdb.From(models.ObjectTag{}).
		Where("post_id=? AND tag_id=?", pid, tagObj.ID).
		MustDelete()
}

func (s *Service) deletePostTags(ctx context.Context, postID int64) {
	s.tdb.From(models.ObjectTag{}).Where(`post_id=?`, postID).MustDelete()
}

func (s *Service) addObjectTag(pid int64, tid int64) {
	objtag := models.ObjectTag{
		PostID: pid,
		TagID:  tid,
	}
	err := s.tdb.Model(&objtag).Create()
	if err == nil {
		return
	}
	if _, ok := err.(*taorm.DupKeyError); ok {
		return
	}
	panic(err)
}

func (s *Service) hasTagName(tagName string) bool {
	var tag models.Tag
	err := s.tags().Where("name=?", tagName).Find(&tag)
	if err == nil {
		return true
	}
	if taorm.IsNotFoundError(err) {
		return false
	}
	panic(err)
}

func (s *Service) addTag(tagName string) int64 {
	tagObj := models.Tag{
		Name: tagName,
	}
	s.tdb.Model(&tagObj).MustCreate()
	return tagObj.ID
}

func (s *Service) getRootTag(tagName string) models.Tag {
	tagObj := s.GetTagByName(tagName)
	if tagObj.Alias == 0 {
		return *tagObj
	}
	ID := tagObj.Alias
	for {
		var tagObj models.Tag
		s.tdb.Where("id=?", ID).MustFind(&tagObj)
		if tagObj.Alias == 0 {
			return tagObj
		}
		ID = tagObj.Alias
	}
}
