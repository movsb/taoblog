package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/db"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
)

func (s *Service) tags() *taorm.Stmt {
	return s.tdb.Model(models.Tag{})
}

// GetTagByName gets a tag by Name.
func (s *Service) GetTagByName(name string) *models.Tag {
	var tag models.Tag
	err := s.tags().Where("name=?", name).Find(&tag)
	if err != nil {
		panic(err)
	}
	return &tag
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

func (s *Service) GetObjectTagNames(postID int64) []string {
	query := `select tags.name from post_tags,tags where post_tags.post_id=? and post_tags.tag_id=tags.id`
	args := []any{postID}
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
	if len(ids) <= 0 {
		return ids
	}
	sids := utils.Join(ids, `,`)

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

// 会自动去重。
func (s *Service) updateObjectTags(pid int64, tags []string) {
	slices.Sort(tags)
	newTags := slices.Compact(tags)
	oldTags := s.GetObjectTagNames(pid)

	var (
		toBeDeled []string
		toBeAdded []string
	)

	for _, t := range oldTags {
		if !slices.ContainsFunc(newTags, func(old string) bool {
			return strings.EqualFold(old, t)
		}) {
			toBeDeled = append(toBeDeled, t)
		}
	}

	for _, t := range newTags {
		t = strings.TrimSpace(t)
		if t != "" && !slices.ContainsFunc(oldTags, func(new string) bool {
			return strings.EqualFold(new, t)
		}) {
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
			tid = s.GetTagByName(t).ID
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

////////////////////////////////////////////////////////////////////////////////

func (s *Service) CreateCategory(ctx context.Context, in *proto.Category) (_ *proto.Category, outErr error) {
	defer utils.CatchAsError(&outErr)

	if s := strings.TrimSpace(in.Name); s == `` || len(in.Name) > 32 {
		panic(`分类名字太短或太长。`)
	}

	ac := user.MustNotBeGuest(ctx)

	cat := models.Category{
		UserID: int32(ac.User.ID),
		Name:   in.Name,
	}

	db := db.FromContextDefault(ctx, s.tdb)
	db.Model(&cat).MustCreate()

	return cat.ToProto()
}

func (s *Service) UpdateCategory(ctx context.Context, in *proto.UpdateCategoryRequest) (_ *proto.Category, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)

	m := map[string]any{}

	if in.UpdateName {
		if s := strings.TrimSpace(in.Category.Name); s == `` || len(in.Category.Name) > 32 {
			panic(`分类名字太短或太长。`)
		}
		m[`name`] = in.Category.Name
	}

	cat := utils.Must1(s.getCatByID(ctx, in.Category.Id))
	if cat.UserID != int32(ac.User.ID) {
		panic(noPerm)
	}

	db := db.FromContextDefault(ctx, s.tdb)
	r := db.Model(cat).MustUpdateMap(m)
	if n, err := r.RowsAffected(); err != nil || n != 1 {
		panic(fmt.Errorf(`更新分类失败：%v, n=%d`, err, n))
	}

	return utils.Must1(s.getCatByID(ctx, in.Category.Id)).ToProto()
}

func (s *Service) ListCategories(ctx context.Context, in *proto.ListCategoriesRequest) (_ *proto.ListCategoriesResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)
	db := db.FromContextDefault(ctx, s.tdb)

	var cats models.Categories
	db.Where(`user_id=?`, ac.User.ID).OrderBy(`id asc`).MustFind(&cats)

	out := utils.Must1(cats.ToProto())
	return &proto.ListCategoriesResponse{
		Categories: out,
	}, nil
}

func (s *Service) getCatByID(ctx context.Context, id int32) (*models.Category, error) {
	var cat models.Category
	db := db.FromContextDefault(ctx, s.tdb)
	return &cat, db.Where(`id=?`, id).Find(&cat)
}

// 检查文章所属的分类是否正确/合法。
func (s *Service) checkPostCat(ctx context.Context, id int32) error {
	// 未分类/默认分类。
	if id == 0 {
		return nil
	}

	cat, err := s.getCatByID(ctx, id)
	if err != nil {
		return fmt.Errorf(`获取分类失败：%w`, err)
	}

	ac := user.Context(ctx)
	if ac.User.ID != int64(cat.UserID) {
		return fmt.Errorf(`获取分类失败：分类不存在`)
	}

	return nil
}
