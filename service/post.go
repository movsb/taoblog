package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/movsb/taoblog/modules/exception"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/post_translators"
	"github.com/movsb/taorm/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	postUntitled = `Untitled`
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
	user := s.auth.User(ctx)
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

// GetPost ...
func (s *Service) GetPost(ctx context.Context, in *protocols.GetPostRequest) (*protocols.Post, error) {
	user := s.auth.AuthGRPC(ctx)

	var p models.Post
	if err := s.tdb.Where("id = ?", in.Id).Find(&p); err != nil {
		if taorm.IsNotFoundError(err) {
			panic(status.Error(codes.NotFound, `post not found`))
		}
		panic(err)
	}

	if p.Status != `public` && !user.IsAdmin() {
		panic(status.Error(codes.NotFound, `post not found`))
	}

	out := p.ToProtocols()

	// TODO don't get these fields
	if !in.WithContent {
		out.Content = ``
	}
	if !in.WithSource {
		out.SourceType = ``
		out.Source = ``
	}
	if in.WithTags {
		out.Tags = s.GetPostTags(out.Id)
	}

	return out, nil
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
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=?"
	s.tdb.MustExec(query, id)
}

func (s *Service) GetPostByPage(parents string, slug string) *protocols.Post {
	return s.mustGetPostBySlugOrPage(true, parents, slug)
}

func (s *Service) GetPostBySlug(categories string, slug string) *protocols.Post {
	return s.mustGetPostBySlugOrPage(false, categories, slug)
}

// GetPostsByTags gets tag posts.
func (s *Service) GetPostsByTags(tags []string) models.Posts {
	var ids []int64
	for _, tag := range tags {
		t := s.GetTagByName(tag)
		ids = append(ids, t.ID)
	}
	tagIDs := s.getAliasTagsAll(ids)
	var posts models.Posts
	s.tdb.From(models.Post{}).From(models.ObjectTag{}).Select("posts.id,posts.title").
		Where("posts.id=post_tags.post_id").
		Where("post_tags.tag_id in (?)", tagIDs).
		GroupBy(`posts.id`).
		Having(fmt.Sprintf(`COUNT(posts.id) >= %d`, len(ids))).
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
// e.g. /folder/post.html, then tree is /folder
// e.g. /path/to/folder/post.html, then tree is /path/to/folder
// It will get the ID of folder.
func (s *Service) parseCategoryTree(tree string) (id int64) {
	if tree == "" {
		return 0
	}

	p := strings.LastIndexByte(tree, '/')
	if p == -1 {
		panic(`invalid tree`)
	}
	path, slug := tree[:p], tree[p+1:]
	if path == `` {
		path = `/`
	}

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

func (s *Service) UpdatePostCommentCount(name int64) {
	var count uint
	s.tdb.Model(models.Comment{}).Where(`post_id=?`, name).MustCount(&count)
	s.tdb.MustExec(`UPDATE posts SET comments=? WHERE id=?`, count, name)
}

// CreatePost ...
func (s *Service) CreatePost(ctx context.Context, in *protocols.Post) (*protocols.Post, error) {
	var err error

	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, `not enough permission`)
	}

	createdAt := int32(time.Now().Unix())

	p := models.Post{
		Date:       createdAt,
		Modified:   createdAt,
		Title:      strings.TrimSpace(in.Title),
		Slug:       in.Slug,
		Type:       in.Type,
		Category:   0,
		Status:     "draft",
		Metas:      "{}",
		Source:     in.Source,
		SourceType: in.SourceType,
	}

	if p.Type == `` {
		p.Type = `post`
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

	cb := post_translators.Callback{
		SetTitle: func(title string) {
			p.Title = title
		},
	}

	// TODO doesn't exist
	p.Content, err = tr.Translate(&cb, in.Source, "./files/0")
	if err != nil {
		panic(err)
	}

	if p.Title == "" {
		p.Title = postUntitled
	}

	s.TxCall(func(txs *Service) error {
		txs.tdb.Model(&p).MustCreate()
		in.Id = p.ID
		txs.UpdateObjectTags(p.ID, in.Tags)
		txs.updateLastPostTime(time.Unix(int64(p.Modified), 0))
		txs.updatePostPageCount()
		return nil
	})

	return p.ToProtocols(), nil
}

// UpdatePost ...
func (s *Service) UpdatePost(ctx context.Context, in *protocols.UpdatePostRequest) (*protocols.Post, error) {
	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, `not enough permission`)
	}

	if in.Post == nil || in.Post.Id == 0 || in.UpdateMask == nil {
		panic(exception.NewValidationError("无效文章"))
	}

	p := models.Post{
		ID: in.Post.Id,
	}

	modified := time.Now().Unix()

	m := map[string]interface{}{
		`modified`: modified,
	}

	var hasSourceType, hasSource bool
	var hasTags bool

	for _, path := range in.UpdateMask.Paths {
		switch path {
		case `title`:
			m[path] = in.Post.Title
			if in.Post.Title == `` {
				m[path] = postUntitled
			}
		case `source_type`:
			m[path] = in.Post.SourceType
			hasSourceType = true
		case `source`:
			m[path] = in.Post.Source
			hasSource = true
		case `content`:
			m[path] = in.Post.Content
		case `slug`:
			m[path] = in.Post.Slug
		case `tags`:
			hasTags = true
		default:
			panic(`unknown update mask`)
		}
	}

	if hasSourceType != hasSource {
		panic(`source type and source must be specified`)
	}

	if hasSource && hasSourceType {
		var tr post_translators.PostTranslator
		switch in.Post.SourceType {
		case "html":
			tr = &post_translators.HTMLTranslator{}
		case "markdown":
			tr = &post_translators.MarkdownTranslator{}
		default:
			panic("no translator found")
		}
		cb := post_translators.Callback{
			SetTitle: func(title string) {
				if title == `` {
					title = postUntitled
				}
				m[`title`] = title
			},
		}

		content, err := tr.Translate(&cb, in.Post.Source, fmt.Sprintf("./files/%d", in.Post.Id))
		if err != nil {
			panic(err)
		}

		m[`content`] = content
	}

	s.TxCall(func(txs *Service) error {
		res := txs.tdb.Model(p).Where(`modified=?`, in.Post.Modified).MustUpdateMap(m)
		rowsAffected, err := res.RowsAffected()
		if err != nil || rowsAffected != 1 {
			op := models.Post{ID: in.Post.Id}
			txs.tdb.Model(&op).MustFind(&op)
			return fmt.Errorf("update failed, modified conflict: %v (modified: %v)", err, op.Modified)
		}
		if hasTags {
			txs.UpdateObjectTags(p.ID, in.Post.Tags)
		}
		txs.updateLastPostTime(time.Now())
		return nil
	})

	np := s.GetPostByID(p.ID)
	if hasTags {
		np.Tags = s.GetPostTags(p.ID)
	}

	return np, nil
}

// DeletePost ...
func (s *Service) DeletePost(ctx context.Context, in *protocols.DeletePostRequest) (*empty.Empty, error) {
	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, `not enough permission`)
	}

	s.TxCall(func(txs *Service) error {
		var p models.Post
		txs.tdb.Select(`id`).Where(`id=?`, in.Id).MustFind(&p)
		txs.tdb.Model(&p).MustDelete()
		txs.deletePostComments(ctx, int64(in.Id))
		txs.deletePostTags(ctx, int64(in.Id))
		txs.updatePostPageCount()
		txs.updateCommentsCount()
		return nil
	})
	return &empty.Empty{}, nil
}

// updateLastPostTime updates last_post_time in options.
func (s *Service) updateLastPostTime(t time.Time) {
	s.SetOption("last_post_time", t.Unix())
}

func (s *Service) updatePostPageCount() {
	var postCount, pageCount int
	s.tdb.Model(models.Post{}).Select(`count(1) as count`).Where(`type='post'`).MustFind(&postCount)
	s.tdb.Model(models.Post{}).Select(`count(1) as count`).Where(`type='page'`).MustFind(&pageCount)
	s.SetOption(`post_count`, postCount)
	s.SetOption(`page_count`, pageCount)
}

// SetPostStatus sets post status.
func (s *Service) SetPostStatus(ctx context.Context, in *protocols.SetPostStatusRequest) (*protocols.SetPostStatusResponse, error) {
	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() {
		return nil, status.Error(codes.PermissionDenied, `not enough permission`)
	}

	s.TxCall(func(txs *Service) error {
		var post models.Post
		txs.tdb.Select("id").Where("id=?", in.Id).MustFind(&post)

		status := `public`
		if !in.Public {
			status = `draft`
		}

		m := map[string]interface{}{
			"status": status,
		}

		if in.Touch {
			now := time.Now().Unix()
			m[`date`] = now
			m[`modified`] = now
		}

		txs.tdb.Model(&post).MustUpdateMap(m)
		return nil
	})
	return &protocols.SetPostStatusResponse{}, nil
}

// GetPostSource ...
func (s *Service) GetPostSource(ctx context.Context, in *protocols.GetPostSourceRequest) (*protocols.GetPostSourceResponse, error) {
	var p models.Post
	s.tdb.Select(`status,source,source_type,content`).Where(`id=?`, in.Id).MustFind(&p)
	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() && p.Status != `public` {
		return nil, status.Error(codes.PermissionDenied, `not enough permission`)
	}
	rsp := &protocols.GetPostSourceResponse{
		Type: p.SourceType,
	}
	switch {
	case p.Source != ``:
		rsp.Content = p.Source
	default:
		rsp.Content = p.Content
	}
	return rsp, nil
}

// GetPostCommentsCount ...
func (s *Service) GetPostCommentsCount(ctx context.Context, in *protocols.GetPostCommentsCountRequest) (*protocols.GetPostCommentsCountResponse, error) {
	var post models.Post
	s.posts().Select("comments").Where("id=?", in.PostId).MustFind(&post)
	return &protocols.GetPostCommentsCountResponse{
		Count: int64(post.Comments),
	}, nil
}
