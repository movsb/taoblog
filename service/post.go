package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type _PostContentCacheKey struct {
	ID      int64
	Options string
}

const (
	postUntitled = `Untitled`
)

func (s *Service) posts() *taorm.Stmt {
	return s.tdb.Model(models.Post{})
}

// MustGetPost ...
func (s *Service) MustGetPost(ctx context.Context, name int64, co *protocols.PostContentOptions) *protocols.Post {
	var p models.Post
	stmt := s.posts().Where("id = ?", name)
	stmt.MustFind(&p)
	return utils.Must(p.ToProtocols(s.setPostExtraFields(ctx, co)))
}

// 奇怪，为什么这个方法不是 protobuf 写的？🤔
func (s *Service) MustListPosts(ctx context.Context, in *protocols.ListPostsRequest) []*protocols.Post {
	ac := auth.Context(ctx)

	var posts models.Posts
	stmt := s.posts().Select(in.Fields).Limit(in.Limit).OrderBy(in.OrderBy)
	stmt.WhereIf(ac.User.IsGuest(), "status = 'public'")
	// 为了兼容，并不能列举碎碎念。
	if in.Kind == "" {
		stmt.Where(`(type=? OR type=?)`, `post`, `page`)
	} else {
		panic("不支持其它种类。")
	}
	if err := stmt.Find(&posts); err != nil {
		panic(err)
	}

	return utils.Must(posts.ToProtocols(s.setPostExtraFields(ctx, &in.ContentOptions)))
}

// TODO：改成 grpc。
func (s *Service) MustListLatestTweets(ctx context.Context, co *protocols.PostContentOptions) []*protocols.Post {
	ac := auth.Context(ctx)

	stmt := s.tdb.Select("*").OrderBy(`date desc`)
	stmt.WhereIf(ac.User.IsGuest(), "status = 'public'")
	stmt.Where(`type=?`, `tweet`)
	// TODO 乱写的。
	stmt.Limit(30000)

	var posts models.Posts
	if err := stmt.Find(&posts); err != nil {
		panic(err)
	}

	return utils.Must(posts.ToProtocols(s.setPostExtraFields(ctx, co)))
}

// TODO 性能很差！
func (s *Service) isPostPublic(ctx context.Context, id int64) bool {
	p := s.MustGetPost(ctx, id, nil)
	return p.Status == `public`
}

// 获取指定编号的文章。
//
// NOTE：如果是公开文章但是非管理员用户，会过滤掉敏感字段。
func (s *Service) GetPost(ctx context.Context, in *protocols.GetPostRequest) (*protocols.Post, error) {
	ac := auth.Context(ctx)

	var p models.Post
	if err := s.tdb.Where("id = ?", in.Id).Find(&p); err != nil {
		panic(err)
	}

	if p.Status != `public` && !ac.User.IsAdmin() {
		panic(codes.PermissionDenied)
	}

	out, err := p.ToProtocols(s.setPostExtraFields(ctx, &protocols.PostContentOptions{
		WithContent:      in.WithContent,
		RenderCodeBlocks: true,
	}))
	if err != nil {
		return nil, err
	}

	if !in.WithSource {
		out.SourceType = ``
		out.Source = ``
	}
	if in.WithTags {
		out.Tags = s.GetPostTags(out.Id)
	}
	if !in.WithMetas {
		out.Metas = nil
	}

	return out, nil
}

func (s *Service) PathResolver(id int64) renderers.PathResolver {
	return &PathResolver{
		base: s.cfg.Data.File.Path,
		fs:   storage.NewLocal(s.cfg.Data.File.Path, id),
	}
}

type PathResolver struct {
	base string
	fs   storage.FileSystem
}

// 1.jpg -> files/{id}/1.jpg
// /{Id}/1.jpg -> files/{id}/1.jpg
func (r *PathResolver) Resolve(path string) string {
	if strings.HasPrefix(path, `/`) {
		return filepath.Join(r.base, path)
	}
	return r.fs.Resolve(path)
}

func (s *Service) getPostContentCached(ctx context.Context, id int64, co *protocols.PostContentOptions) (string, error) {
	key := _PostContentCacheKey{
		ID:      id,
		Options: co.String(),
	}
	content, err, ok := s.postContentCaches.GetOrLoad(ctx, key,
		func(ctx context.Context, key _PostContentCacheKey) (string, time.Duration, error) {
			content, err := s.getPostContent(id, co)
			if err != nil {
				return ``, 0, err
			}
			s.postCaches.Append(id, key)
			return content, time.Minute * 10, nil
		},
	)
	if err != nil {
		return ``, err
	}
	log.Println(`cache status: get_post_content:`, ok, key)
	return content, nil
}

func (s *Service) deletePostContentCacheFor(id int64) {
	log.Println(`即将删除文章缓存：`, id)
	s.postCaches.Delete(id, func(second _PostContentCacheKey) {
		s.postContentCaches.Delete(second)
		log.Println(`删除文章缓存：`, second)
	})
}

func (s *Service) getPostContent(id int64, co *protocols.PostContentOptions) (string, error) {
	if !co.WithContent {
		panic(`without content but get content`)
	}
	var p models.Post
	if err := s.tdb.Select("type,source_type,source,metas").Where("id = ?", id).Find(&p); err != nil {
		return "", err
	}
	var content string
	var tr renderers.Renderer
	switch p.SourceType {
	case `markdown`:
		options := []renderers.Option2{
			renderers.WithPathResolver(s.PathResolver(id)),
			renderers.WithRemoveTitleHeading(true),
			renderers.WithAssetSources(func(path string) (name string, url string, description string, found bool) {
				if src, ok := p.Metas.Sources[path]; ok {
					name = src.Name
					url = src.URL
					description = src.Description
					found = true
				}
				return
			}),
			renderers.WithOpenLinksInNewTab(renderers.OpenLinksInNewTabKindExternal),
		}
		if link := s.GetLink(id); link != s.plainLink(id) {
			options = append(options, renderers.WithModifiedAnchorReference(link))
		}
		if co.UseAbsolutePaths {
			options = append(options, renderers.WithUseAbsolutePaths(s.plainLink(id)))
		}
		if co.RenderCodeBlocks {
			options = append(options, renderers.WithRenderCodeAsHTML())
		}
		tr = renderers.NewMarkdown(options...)
	case `html`:
		tr = &renderers.HTML{}
	default:
		return ``, fmt.Errorf(`unknown source type`)
	}
	_, content, err := tr.Render(p.Source)
	return content, err
}

func (s *Service) GetPostTitle(ID int64) string {
	var p models.Post
	s.posts().Select("title").Where("id = ?", ID).MustFind(&p)
	return p.Title
}

// 临时放这儿。
// 本应该由各主题自己实现的。
func (s *Service) GetLink(ID int64) string {
	var p models.Post
	s.posts().Select("id,slug,category,type").Where("id = ?", ID).MustFind(&p)
	if p.Type == `page` && p.Slug != "" && p.Category == 0 {
		return fmt.Sprintf(`/%s`, p.Slug)
	}
	return s.plainLink(p.ID)
}

// 普通链接是为了附件的 <base> 而设置，对任何主题都生效。
func (s *Service) GetPlainLink(id int64) string {
	return s.plainLink(id)
}
func (s *Service) plainLink(id int64) string {
	return fmt.Sprintf(`/%d/`, id)
}

func (s *Service) GetPostByID(ctx context.Context, id int64, co *protocols.PostContentOptions) *protocols.Post {
	return s.MustGetPost(ctx, id, co)
}

func (s *Service) IncrementPostPageView(id int64) {
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=?"
	s.tdb.MustExec(query, id)
}

func (s *Service) GetPostByPage(ctx context.Context, parents string, slug string, co *protocols.PostContentOptions) *protocols.Post {
	return s.mustGetPostBySlugOrPage(ctx, true, parents, slug, co)
}

func (s *Service) GetPostBySlug(ctx context.Context, categories string, slug string, co *protocols.PostContentOptions) *protocols.Post {
	return s.mustGetPostBySlugOrPage(ctx, false, categories, slug, co)
}

// GetPostsByTags gets tag posts.
func (s *Service) GetPostsByTags(ctx context.Context, req *protocols.GetPostsByTagsRequest) (*protocols.GetPostsByTagsResponse, error) {
	var ids []int64
	for _, tag := range req.Tags {
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
	outs, err := posts.ToProtocols(s.setPostExtraFields(ctx, req.ContentOptions))
	if err != nil {
		return nil, err
	}
	return &protocols.GetPostsByTagsResponse{Posts: outs}, nil
}

func (s *Service) mustGetPostBySlugOrPage(ctx context.Context, isPage bool, parents string, slug string, co *protocols.PostContentOptions) *protocols.Post {
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
		panic(err)
	}
	return utils.Must(p.ToProtocols(s.setPostExtraFields(ctx, co)))
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
		panic(err)
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

// t: last_commented_at 表示文章评论最后被操作的时间。不是最后被评论的时间。
// 因为属于是外部关联资源，对 304 有贡献。
func (s *Service) updatePostCommentCount(pid int64, t time.Time) {
	var count uint
	s.tdb.Model(models.Comment{}).Where(`post_id=?`, pid).MustCount(&count)
	s.tdb.MustExec(`UPDATE posts SET comments=?,last_commented_at=? WHERE id=?`, count, t.Unix(), pid)
}

// CreatePost ...
func (s *Service) CreatePost(ctx context.Context, in *protocols.Post) (*protocols.Post, error) {
	s.MustBeAdmin(ctx)

	now := int32(time.Now().Unix())

	p := models.Post{
		ID:         in.Id,
		Date:       0,
		Modified:   0,
		Title:      strings.TrimSpace(in.Title),
		Slug:       in.Slug,
		Type:       in.Type,
		Category:   0,
		Status:     "draft",
		Metas:      *models.PostMetaFrom(in.Metas),
		Source:     in.Source,
		SourceType: in.SourceType,
	}

	if strings.TrimSpace(p.Source) == "" {
		return nil, status.Error(codes.InvalidArgument, "内容不应为空。")
	}

	if in.Modified > 0 {
		p.Modified = in.Modified
	}
	if in.Date > 0 {
		p.Date = in.Date
		if in.Modified == 0 {
			p.Modified = p.Date
		}
	} else {
		p.Date = now
		p.Modified = now
	}

	if in.Status != "" {
		p.Status = in.Status
	}

	if p.Type == `` {
		p.Type = `post`
	}

	if p.Title == "" {
		p.Title = postUntitled
	}

	s.MustTxCall(func(txs *Service) error {
		txs.tdb.Model(&p).MustCreate()
		in.Id = p.ID
		txs.UpdateObjectTags(p.ID, in.Tags)
		txs.updateLastPostTime(time.Unix(int64(p.Modified), 0))
		txs.updatePostPageCount()
		return nil
	})

	// TODO 暂时没提供选项。
	return p.ToProtocols(s.setPostExtraFields(ctx, nil))
}

// UpdatePost ...
func (s *Service) UpdatePost(ctx context.Context, in *protocols.UpdatePostRequest) (*protocols.Post, error) {
	s.MustBeAdmin(ctx)

	if in.Post == nil || in.Post.Id == 0 || in.UpdateMask == nil {
		return nil, status.Error(codes.InvalidArgument, "无效文章编号、更新字段")
	}

	p := models.Post{
		ID: in.Post.Id,
	}

	now := time.Now().Unix()

	m := map[string]interface{}{
		`modified`: now,
	}

	var hasSourceType, hasSource bool
	var hasTags, hasMetas bool

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
		case `slug`:
			m[path] = in.Post.Slug
		case `tags`:
			hasTags = true
		case `metas`:
			hasMetas = true
		case `type`:
			m[path] = in.Post.Type
		case `created`:
			m[`date`] = in.Post.Date
		default:
			panic(`unknown update mask:` + path)
		}
	}

	if hasSourceType != hasSource {
		panic(`source type and source must be specified`)
	}

	if hasSource && hasSourceType {
		var tr renderers.Renderer
		switch in.Post.SourceType {
		case "html":
			tr = &renderers.HTML{}
		case "markdown":
			// 这里只是用 title 的话，可以不考虑 Markdown 的初始化参数。
			tr = renderers.NewMarkdown()
		default:
			panic("no renderer was found")
		}

		title, _, err := tr.Render(in.Post.Source)
		if err != nil {
			panic(err)
		}

		// 有些旧文章 MD 没有标题，标题在 config 里面，此处不能强制替换。
		if title != `` {
			m[`title`] = title
		}
	}
	if hasMetas {
		m[`metas`] = models.PostMetaFrom(in.Post.Metas)
	}

	s.MustTxCall(func(txs *Service) error {
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
		txs.deletePostContentCacheFor(p.ID)
		return nil
	})

	// TODO 暂时没提供选项。
	np := s.GetPostByID(ctx, p.ID, nil)
	if hasTags {
		np.Tags = s.GetPostTags(p.ID)
	}

	return np, nil
}

// 用于删除一篇文章。
// 这个函数基本没怎么测试过，因为基本上只是设置为不公开。
func (s *Service) DeletePost(ctx context.Context, in *protocols.DeletePostRequest) (*empty.Empty, error) {
	s.MustBeAdmin(ctx)

	s.MustTxCall(func(txs *Service) error {
		var p models.Post
		txs.tdb.Select(`id`).Where(`id=?`, in.Id).MustFind(&p)
		txs.tdb.Model(&p).MustDelete()
		txs.deletePostComments(ctx, int64(in.Id))
		txs.deletePostTags(ctx, int64(in.Id))
		txs.deletePostContentCacheFor(int64(in.Id))
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
	s.MustBeAdmin(ctx)

	s.MustTxCall(func(txs *Service) error {
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

// GetPostCommentsCount ...
func (s *Service) GetPostCommentsCount(ctx context.Context, in *protocols.GetPostCommentsCountRequest) (*protocols.GetPostCommentsCountResponse, error) {
	var post models.Post
	s.posts().Select("comments").Where("id=?", in.PostId).MustFind(&post)
	return &protocols.GetPostCommentsCountResponse{
		Count: int64(post.Comments),
	}, nil
}

func (s *Service) FindRedirect(sourcePath string) (string, error) {
	r, err := s.findRedirect(context.TODO(), s, sourcePath)
	if err != nil {
		return "", err
	}
	return r.TargetPath, nil
}

func (s *Service) findRedirect(_ context.Context, txs *Service, sourcePath string) (*models.Redirect, error) {
	var r models.Redirect
	if err := txs.tdb.Where(`source_path=?`, sourcePath).Find(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// 重定向并不止是只对文章链接有效，任何链接都可以。这里暂时先写在这里了。
func (s *Service) SetRedirect(ctx context.Context, in *protocols.SetRedirectRequest) (*empty.Empty, error) {
	return &empty.Empty{}, s.TxCall(func(txs *Service) error {
		r, err := s.findRedirect(ctx, txs, in.SourcePath)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				var te *taorm.Error
				if errors.As(err, &te) {
					if !errors.Is(te.Raw, sql.ErrNoRows) {
						return err
					}
				}
			}
		}
		if r != nil {
			if r.TargetPath == in.TargetPath {
				return nil
			}
			res, err := txs.tdb.Model(r).UpdateMap(taorm.M{
				`target_path`: in.TargetPath,
			})
			if err != nil {
				return err
			}
			n, err := res.RowsAffected()
			if err != nil {
				return err
			}
			if n != 1 {
				return fmt.Errorf(`SetRedirect: affected rows not equal to 1, was %d`, n)
			}
			return nil
		}
		r = &models.Redirect{
			CreatedAt:  int32(time.Now().Unix()),
			SourcePath: in.SourcePath,
			TargetPath: in.TargetPath,
		}
		return txs.tdb.Model(r).Create()
	})
}

func (s *Service) setPostExtraFields(ctx context.Context, co *protocols.PostContentOptions) func(c *protocols.Post) error {
	ac := auth.Context(ctx)

	return func(p *protocols.Post) error {
		if !ac.User.IsAdmin() {
			p.Metas.Geo = nil
		}

		if co != nil && co.WithContent {
			content, err := s.getPostContentCached(ctx, p.Id, co)
			if err != nil {
				return err
			}
			p.Content = content
		}
		return nil
	}
}
