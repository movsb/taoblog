package service

import (
	"fmt"
	"strings"

	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) posts() *taorm.Stmt {
	return s.tdb.Model(models.Post{}, "posts")
}

// GetPost ...
func (s *ImplServer) GetPost(name int64) *models.Post {
	var post models.Post
	s.posts().Where("id = ?", name).Find(&post)
	return &post
}

// ListPosts ...
func (s *ImplServer) ListPosts(in *ListPostsRequest) []*models.Post {
	var posts []*models.Post
	s.posts().Select(in.Fields).Limit(in.Limit).OrderBy(in.OrderBy).Find(&posts)
	return posts
}

func (s *ImplServer) GetLatestPosts(fields string, limit int64) []*models.Post {
	in := ListPostsRequest{
		Fields:  fields,
		Limit:   limit,
		OrderBy: "date DESC",
	}
	return s.ListPosts(&in)
}

func (s *ImplServer) GetPostTitle(ID int64) string {
	var p models.Post
	s.posts().Select("title").Where("id = ?", ID).Find(&p)
	return p.Title
}

func (s *ImplServer) GetPostByID(id int64) *models.Post {
	return s.GetPost(id)
}

func (s *ImplServer) IncrementPostPageView(id int64) {
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=? LIMIT 1"
	s.tdb.MustExec(query, id)
}

func (s *ImplServer) GetPostByPage(parents string, slug string) *models.Post {
	return s.getPostBySlugOrPage(true, parents, slug)
}

func (s *ImplServer) GetPostBySlug(categories string, slug string) *models.Post {
	return s.getPostBySlugOrPage(false, categories, slug)
}

// GetPostsByTags gets tag posts.
func (s *ImplServer) GetPostsByTags(tagName string) []*models.PostForArchive {
	tag := s.GetTagByName(tagName)
	tagIDs := s.getAliasTagsAll([]int64{tag.ID})
	var posts []*models.PostForArchive
	s.tdb.From("posts").From("post_tags").Select("posts.id,posts.title").
		Where("posts.id=post_tags.post_id").
		Where("post_tags.tag_id in (?)", tagIDs).
		Find(&posts)
	return posts
}

func (s *ImplServer) getPostBySlugOrPage(isPage bool, parents string, slug string) *models.Post {
	var catID int64
	if !isPage {
		catID = s.parseCategoryTree(parents)
	} else {
		catID = s.getPageParentID(parents)
	}
	var p models.Post
	s.tdb.Model(models.Post{}, "posts").
		Where("slug=? AND taxonomy=?", slug, catID).
		WhereIf(isPage, "type = 'page'").
		OrderBy("date DESC").
		Find(&p)
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
	var cats []*models.Category
	s.tdb.Model(models.Post{}, "posts").Where("slug IN (?)", parts).Find(&cats)
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

	type getPageParentID_Result struct {
		ID     int64
		Slug   string
		Parent int64
	}

	var results []*getPageParentID_Result
	s.tdb.Model(models.Post{}, "posts").
		Select("id,slug,taxonomy").
		Where("slug IN (?)", slugs).
		Where("type = 'page'").
		Find(&results)
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

func (s *ImplServer) GetRelatedPosts(id int64) []*models.PostForRelated {
	tagIDs := s.getObjectTagIDs(id, true)
	if len(tagIDs) == 0 {
		return make([]*models.PostForRelated, 0)
	}
	var relates []*models.PostForRelated
	s.tdb.From("posts").From("post_tags").
		Select("posts.id,posts.title,COUNT(posts.id) relevance").
		Where("post_tags.post_id != ?", id).
		Where("posts.id = post_tags.post_id").
		Where("post_tags.tag_id IN (?)", tagIDs).
		GroupBy("posts.id").
		OrderBy("relevance DESC").
		Limit(9).
		Find(&relates)
	return relates
}

func (s *ImplServer) GetPostTags(ID int64) []string {
	return s.GetObjectTagNames(ID)
}
