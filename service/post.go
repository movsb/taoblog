package service

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/post_translators"
)

func (s *ImplServer) posts() *taorm.Stmt {
	return s.tdb.Model(models.Post{}, "posts")
}

// MustGetPost ...
func (s *ImplServer) MustGetPost(name int64) *protocols.Post {
	var p models.Post
	stmt := s.posts().Where("id = ?", name)
	if err := stmt.Find(&p); err != nil {
		if err == sql.ErrNoRows {
			panic(&PostNotFoundError{})
		}
		panic(err)
	}

	out := p.ToProtocols()

	return out
}

// MustListPosts ...
func (s *ImplServer) MustListPosts(in *protocols.ListPostsRequest) []*protocols.Post {
	var posts models.Posts
	stmt := s.posts().Select(in.Fields).Limit(in.Limit).OrderBy(in.OrderBy)
	if err := stmt.Find(&posts); err != nil {
		panic(err)
	}
	return posts.ToProtocols()
}

func (s *ImplServer) GetLatestPosts(fields string, limit int64) []*protocols.Post {
	in := protocols.ListPostsRequest{
		Fields:  fields,
		Limit:   limit,
		OrderBy: "date DESC",
	}
	return s.MustListPosts(&in)
}

func (s *ImplServer) GetPostTitle(ID int64) string {
	var p models.Post
	s.posts().Select("title").Where("id = ?", ID).MustFind(&p)
	return p.Title
}

func (s *ImplServer) GetPostByID(id int64) *protocols.Post {
	return s.MustGetPost(id)
}

func (s *ImplServer) IncrementPostPageView(id int64) {
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=? LIMIT 1"
	s.tdb.MustExec(query, id)
}

func (s *ImplServer) GetPostByPage(parents string, slug string) *protocols.Post {
	return s.mustGetPostBySlugOrPage(true, parents, slug)
}

func (s *ImplServer) GetPostBySlug(categories string, slug string) *protocols.Post {
	return s.mustGetPostBySlugOrPage(false, categories, slug)
}

// GetPostsByTags gets tag posts.
func (s *ImplServer) GetPostsByTags(tagName string) []*models.PostForArchive {
	tag := s.GetTagByName(tagName)
	tagIDs := s.getAliasTagsAll([]int64{tag.ID})
	var posts []*models.PostForArchive
	s.tdb.From("posts").From("post_tags").Select("posts.id,posts.title").
		Where("posts.id=post_tags.post_id").
		Where("post_tags.tag_id in (?)", tagIDs).
		MustFind(&posts)
	return posts
}

func (s *ImplServer) mustGetPostBySlugOrPage(isPage bool, parents string, slug string) *protocols.Post {
	var catID int64
	if !isPage {
		catID = s.parseCategoryTree(parents)
	} else {
		catID = s.getPageParentID(parents)
	}
	var p models.Post
	stmt := s.tdb.Model(models.Post{}, "posts").
		Where("slug=? AND taxonomy=?", slug, catID).
		WhereIf(isPage, "type = 'page'").
		OrderBy("date DESC")
	if err := stmt.Find(&p); err != nil {
		if err == sql.ErrNoRows {
			panic(&PostNotFoundError{})
		}
		panic(err)
	}
	p.Date = datetime.My2Local(p.Date)
	p.Modified = datetime.My2Local(p.Modified)
	// TODO
	// p.Tags, _ = tagmgr.GetObjectTagNames(gdb, p.ID)
	return p.ToProtocols()
}

// ParseTree parses category tree from URL to get last sub-category ID.
// e.g. /path/to/folder/post.html, then tree is path/to/folder
// It will get the ID of folder
func (s *ImplServer) parseCategoryTree(tree string) (id int64) {
	parts := strings.Split(tree, "/")
	var cats []*models.Category
	s.tdb.Model(models.Category{}, "taxonomies").Where("slug IN (?)", parts).MustFind(&cats)
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
		MustFind(&results)
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
		MustFind(&relates)
	return relates
}

func (s *ImplServer) GetPostTags(ID int64) []string {
	return s.GetObjectTagNames(ID)
}

func (s *ImplServer) GetPostCommentCount(name int64) (count int64) {
	var post models.Post
	s.posts().Select("comments").Where("id=?", name).MustFind(&post)
	return int64(post.Comments)
}

func (s *ImplServer) UpdatePostCommentCount(name int64) {
	query := `UPDATE posts INNER JOIN (SELECT count(post_id) count FROM comments WHERE post_id=?) x ON posts.id=? SET posts.comments=x.count`
	s.tdb.MustExec(query, name, name)
}

func (s *ImplServer) ListPostComments(in *protocols.ListCommentsRequest) []*models.Comment {
	return s.ListComments(in)
}

// CreatePost ...
func (s *ImplServer) CreatePost(in *protocols.Post) {
	var err error

	createdAt := datetime.MyGmt()

	p := models.Post{
		Date:       createdAt,
		Modified:   createdAt,
		Title:      strings.TrimSpace(in.Title),
		Slug:       "",
		Type:       protocols.PostTypePost,
		Category:   1,
		Status:     "public",
		Metas:      "{}",
		Source:     in.Source,
		SourceType: in.SourceType,
	}

	if p.Title == "" {
		p.Title = "Untitled"
	}

	var tr post_translators.PostTranslator
	switch p.SourceType {
	case "html":
		tr = &post_translators.HTMLTranslator{}
	case "markdown":
		tr = &post_translators.MarkdownTranslator{}
	default:
		panic("no translator found")
	}

	p.Content, err = tr.Translate(in.Source)
	if err != nil {
		panic(err)
	}

	s.TxCall(func(txs *ImplServer) error {
		txs.tdb.Model(&p, "posts").MustCreate()
		return nil
	})
}

func (s *ImplServer) UpdatePost(p *models.Post) {
	var err error

	if p.ID == 0 {
		panic(exception.NewValidationError("无效文章编号"))
	}

	var tr post_translators.PostTranslator
	switch p.SourceType {
	case "html":
		tr = &post_translators.HTMLTranslator{}
	case "markdown":
		tr = &post_translators.MarkdownTranslator{}
	default:
		panic("no translator found")
	}
	p.Content, err = tr.Translate(p.Source)
	if err != nil {
		panic(err)
	}

	p.Modified = datetime.MyGmt()

	s.tdb.Model(p, "posts").UpdateMap(map[string]interface{}{
		"modified":    p.Modified,
		"source_type": p.SourceType,
		"source":      p.Source,
		"content":     p.Content,
	})
}
