package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/post_translators"
	"github.com/movsb/taorm/taorm"
)

func (s *Service) posts() *taorm.Stmt {
	return s.tdb.Model(models.Post{})
}

// MustGetPost ...
func (s *Service) MustGetPost(name int64) *protocols.Post {
	var p models.Post
	stmt := s.posts().Where("id = ?", name)
	if err := stmt.Find(&p); err != nil {
		if taorm.IsNotFoundError(err) {
			panic(&PostNotFoundError{})
		}
		panic(err)
	}

	out := p.ToProtocols()

	return out
}

// MustListPosts ...
func (s *Service) MustListPosts(ctx context.Context, in *protocols.ListPostsRequest) []*protocols.Post {
	user := s.auth.AuthContext(ctx)
	var posts models.Posts
	stmt := s.posts().Select(in.Fields).Limit(in.Limit).OrderBy(in.OrderBy)
	stmt.WhereIf(user.IsGuest(), "status = 'public'")
	if err := stmt.Find(&posts); err != nil {
		panic(err)
	}
	return posts.ToProtocols()
}

func (s *Service) GetLatestPosts(ctx context.Context, fields string, limit int64) []*protocols.Post {
	in := protocols.ListPostsRequest{
		Fields:  fields,
		Limit:   limit,
		OrderBy: "date DESC",
	}
	return s.MustListPosts(ctx, &in)
}

func (s *Service) GetPostTitle(ID int64) string {
	var p models.Post
	s.posts().Select("title").Where("id = ?", ID).MustFind(&p)
	return p.Title
}

func (s *Service) GetPostByID(id int64) *protocols.Post {
	return s.MustGetPost(id)
}

func (s *Service) IncrementPostPageView(id int64) {
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=? LIMIT 1"
	s.tdb.MustExec(query, id)
}

func (s *Service) GetPostByPage(parents string, slug string) *protocols.Post {
	return s.mustGetPostBySlugOrPage(true, parents, slug)
}

func (s *Service) GetPostBySlug(categories string, slug string) *protocols.Post {
	return s.mustGetPostBySlugOrPage(false, categories, slug)
}

// GetPostsByTags gets tag posts.
func (s *Service) GetPostsByTags(tagName string) []*models.PostForArchive {
	tag := s.GetTagByName(tagName)
	tagIDs := s.getAliasTagsAll([]int64{tag.ID})
	var posts []*models.PostForArchive
	s.tdb.From(models.Post{}).From(models.ObjectTag{}).Select("posts.id,posts.title").
		Where("posts.id=post_tags.post_id").
		Where("post_tags.tag_id in (?)", tagIDs).
		MustFind(&posts)
	return posts
}

func (s *Service) mustGetPostBySlugOrPage(isPage bool, parents string, slug string) *protocols.Post {
	var catID int64
	if !isPage {
		catID = s.parseCategoryTree(parents)
	} else {
		catID = s.getPageParentID(parents)
	}
	var p models.Post
	stmt := s.tdb.Model(models.Post{}).
		Where("slug=? AND category=?", slug, catID).
		WhereIf(isPage, "type = 'page'").
		OrderBy("date DESC")
	if err := stmt.Find(&p); err != nil {
		if taorm.IsNotFoundError(err) {
			panic(&PostNotFoundError{})
		}
		panic(err)
	}
	return p.ToProtocols()
}

// ParseTree parses category tree from URL to get last sub-category ID.
// e.g. /folder/post.html, then tree is folder
// e.g. /path/to/folder/post.html, then tree is path/to/folder
// It will get the ID of folder
var reSplitPathAndSlug = regexp.MustCompile(`^(.*)/([^/]+)$`)

func (s *Service) parseCategoryTree(tree string) (id int64) {
	if tree == "" {
		return 1
	}
	tree = "/" + tree
	matches := reSplitPathAndSlug.FindStringSubmatch(tree)
	if matches == nil {
		panic(&CategoryNotFoundError{})
	}
	path := matches[1]
	if path == "" {
		path = "/"
	}
	slug := matches[2]
	var cat models.Category
	if err := s.tdb.Where("path=? AND slug=?", path, slug).Find(&cat); err != nil {
		panic(&CategoryNotFoundError{})
	}
	return cat.ID
}

func (s *Service) getPageParentID(parents string) int64 {
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
	s.tdb.Model(models.Post{}).
		Select("id,slug,category").
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

func (s *Service) GetRelatedPosts(id int64) []*models.PostForRelated {
	tagIDs := s.getObjectTagIDs(id, true)
	if len(tagIDs) == 0 {
		return make([]*models.PostForRelated, 0)
	}
	var relates []*models.PostForRelated
	s.tdb.From(models.Post{}).From(models.ObjectTag{}).
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

func (s *Service) GetPostTags(ID int64) []string {
	return s.GetObjectTagNames(ID)
}

func (s *Service) GetPostCommentCount(name int64) (count int64) {
	var post models.Post
	s.posts().Select("comments").Where("id=?", name).MustFind(&post)
	return int64(post.Comments)
}

func (s *Service) UpdatePostCommentCount(name int64) {
	query := `UPDATE posts INNER JOIN (SELECT count(post_id) count FROM comments WHERE post_id=?) x ON posts.id=? SET posts.comments=x.count`
	s.tdb.MustExec(query, name, name)
}

// CreatePost ...
func (s *Service) CreatePost(in *protocols.Post) {
	var err error

	createdAt := datetime.MyLocal()

	p := models.Post{
		Date:       createdAt,
		Modified:   createdAt,
		Title:      strings.TrimSpace(in.Title),
		Slug:       in.Slug,
		Type:       protocols.PostTypePost,
		Category:   1,
		Status:     "draft",
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

	// TODO doesn't exist
	p.Content, err = tr.Translate(in.Source, "./files/0")
	if err != nil {
		panic(err)
	}

	s.TxCall(func(txs *Service) error {
		txs.tdb.Model(&p).MustCreate()
		in.ID = p.ID
		txs.UpdateObjectTags(p.ID, in.Tags)
		txs.updateLastPostTime(p.Modified)
		txs.updatePostPageCount()
		return nil
	})
}

func (s *Service) UpdatePost(in *protocols.Post) {
	var err error

	if in.ID == 0 {
		panic(exception.NewValidationError("无效文章编号"))
	}

	p := models.Post{
		ID: in.ID,
	}

	var tr post_translators.PostTranslator
	switch in.SourceType {
	case "html":
		tr = &post_translators.HTMLTranslator{}
	case "markdown":
		tr = &post_translators.MarkdownTranslator{}
	default:
		panic("no translator found")
	}
	content, err := tr.Translate(in.Source, fmt.Sprintf("./files/%d", in.ID))
	if err != nil {
		panic(err)
	}

	modified := datetime.MyLocal()

	s.TxCall(func(txs *Service) error {
		txs.tdb.Model(p).MustUpdateMap(map[string]interface{}{
			"title":       in.Title,
			"modified":    modified,
			"source_type": in.SourceType,
			"source":      in.Source,
			"content":     content,
			"slug":        in.Slug,
		})
		txs.UpdateObjectTags(p.ID, in.Tags)
		txs.updateLastPostTime(modified)
		return nil
	})
}

// updateLastPostTime updates last_post_time in options.
func (s *Service) updateLastPostTime(ts string) {
	if ts == "" {
		ts = datetime.MyLocal()
	}
	s.SetOption("last_post_time", ts)
}

func (s *Service) updatePostPageCount() {
	var postCount, pageCount int
	s.tdb.Model(models.Post{}).Select(`count(1) as count`).Where(`type='post'`).MustFind(&postCount)
	s.tdb.Model(models.Post{}).Select(`count(1) as count`).Where(`type='page'`).MustFind(&pageCount)
	s.SetOption(`post_count`, postCount)
	s.SetOption(`page_count`, pageCount)
}

// SetPostStatus sets post status.
func (s *Service) SetPostStatus(id int64, status string) {
	s.TxCall(func(txs *Service) error {
		var post models.Post
		txs.tdb.Select("id").Where("id=?", id).MustFind(&post)
		txs.tdb.Model(&post).MustUpdateMap(map[string]interface{}{
			"status": status,
		})
		return nil
	})
}
