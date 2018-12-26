package service

import (
	"fmt"
	"strings"

	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/sql_helpers"
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/service/models"
)

// GetPost ...
func (s *ImplServer) GetPost(name int64) *models.Post {
	query := `SELECT * FROM posts WHERE id = ?`
	var post models.Post
	taorm.MustQueryRows(&post, s.db, query, name)
	return &post
}

// ListPosts ...
func (s *ImplServer) ListPosts(in *ListPostsRequest) []*models.Post {
	query := `SELECT * FROM posts`
	var posts []*models.Post
	taorm.MustQueryRows(&posts, s.db, query)
	return posts
}

/*
func (s *ImplServer) ListLatestPosts(in *ListLatestPostsRequest) []*models.PostForLatest {
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,title,type").
		Where("type='post'").
		OrderBy("date DESC").
		Limit(in.Limit).
		SQL()
	var ps []*models.PostForLatest
	taorm.MustQueryRows(&ps, s.db, query, args...)
	return ps
}

*/

func (s *ImplServer) GetPostTitle(ID int64) string {
	var p models.Post
	query := `SELECT title FROM posts WHERE id = ?`
	taorm.MustQueryRows(&p, s.db, query, ID)
	return p.Title
}

func (s *ImplServer) GetPostByID(id int64) *models.Post {
	seldb := sql_helpers.NewSelect().From("posts", "").Select("*").Where("id=?", id)
	query, args := seldb.SQL()
	var p models.Post
	taorm.MustQueryRows(&p, s.db, query, args...)
	p.Date = datetime.My2Local(p.Date)
	p.Modified = datetime.My2Local(p.Modified)
	// TODO tags
	//p.Tags, _ = tagmgr.GetObjectTagNames(gdb, p.ID)
	return &p
}

func (s *ImplServer) IncrementPostView(id int64) {
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=? LIMIT 1"
	s.db.Exec(query, id)
}

func (s *ImplServer) GetPostByPage(parents string, slug string) *models.Post {
	return s.getPostBySlugOrPage(true, parents, slug)
}

func (s *ImplServer) GetPostBySlug(categories string, slug string) *models.Post {
	return s.getPostBySlugOrPage(false, categories, slug)
}

func (s *ImplServer) getPostBySlugOrPage(isPage bool, parents string, slug string) *models.Post {
	var catID int64
	if !isPage {
		catID = s.parseCategoryTree(parents)
	} else {
		catID = s.getPageParentID(parents)
	}
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("*").
		Where("slug=? AND taxonomy=?", slug, catID).
		WhereIf(isPage, "type = 'page'").
		OrderBy("date DESC").
		SQL()
	var p models.Post
	taorm.MustQueryRows(&p, s.db, query, args...)
	p.Date = datetime.My2Local(p.Date)
	p.Modified = datetime.My2Local(p.Modified)
	// TODO
	// p.Tags, _ = tagmgr.GetObjectTagNames(gdb, p.ID)
	return &p
}

// ParseTree parses category tree from URL to get last sub-category ID.
// e.g. /path/to/folder/post.html, then tree is path/to/folder
// It will get the ID of folder
func (s *ImplServer) parseCategoryTree(tree string) (id int64) {
	parts := strings.Split(tree, "/")
	query, args := sql_helpers.NewSelect().From("taxonomies", "").
		Select("*").Where("slug IN (?)", parts).SQL()
	var cats []*models.Category
	taorm.MustQueryRows(&cats, s.db, query, args...)
	var parent int64
	for i := 0; i < len(parts); i++ {
		found := false
		for _, cat := range cats {
			if cat.Parent == parent && cat.Slug == parts[i] {
				parent = cat.ID
				found = true
				break
			}
		}
		if !found {
			panic(fmt.Errorf("找不到分类：%s", parts[i]))
		}
	}
	return parent
}

func (s *ImplServer) getPageParentID(parents string) int64 {
	if len(parents) == 0 {
		return 0
	}
	parents = parents[1:]
	slugs := strings.Split(parents, "/")
	query, args := sql_helpers.NewSelect().From("posts", "").
		Select("id,slug,taxonomy").
		Where("slug IN (?)", slugs).
		Where("type = 'page'").
		SQL()

	type getPageParentID_Result struct {
		ID     int64
		Slug   string
		Parent int64
	}

	var results []*getPageParentID_Result
	taorm.MustQueryRows(&results, s.db, query, args...)

	var parent int64
	for i := 0; i < len(slugs); i++ {
		found := false
		for _, r := range results {
			if r.Parent == parent && r.Slug == slugs[i] {
				parent = r.ID
				found = true
				break
			}
			if !found {
				panic(fmt.Errorf("找不到父页面：%s", slugs[i]))
			}
		}
	}

	return parent
}

func (s *ImplServer) GetRelatedPosts(id int64) []*models.Post {
	tagIDs := s.getObjectTagIDs(id, true)
	if len(tagIDs) == 0 {
		return make([]*models.Post, 0)
	}
	query, args := sql_helpers.NewSelect().
		From("posts", "p").
		From("post_tags", "pt").
		Select("p.id,p.title,COUNT(p.id) relevance").
		Where("pt.post_id != ?", id).
		Where("p.id = pt.post_id").
		Where("pt.tag_id IN (?)", tagIDs).
		GroupBy("p.id").
		OrderBy("relevance DESC").
		Limit(9).
		SQL()
	var relates []*models.Post
	taorm.MustQueryRows(&relates, s.db, query, args...)
	return relates
}

func (s *ImplServer) GetPostTags(ID int64) []string {
	return s.getObjectTagNames(ID)
}
