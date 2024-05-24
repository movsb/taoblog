package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	proto "github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/media_tags"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type _PostContentCacheKey struct {
	ID      int64
	Options string
}

func (s *Service) posts() *taorm.Stmt {
	return s.tdb.Model(models.Post{})
}

func (s *Service) ListPosts(ctx context.Context, in *proto.ListPostsRequest) (*proto.ListPostsResponse, error) {
	ac := auth.Context(ctx)

	var posts models.Posts
	stmt := s.posts().Limit(int64(in.Limit)).OrderBy(in.OrderBy)

	stmt.WhereIf(ac.User.IsGuest(), "status = 'public'")
	stmt.WhereIf(len(in.Kinds) > 0, `type in ?`, in.Kinds)
	stmt.WhereIf(in.ModifiedNotBefore > 0, `modified >= ?`, in.ModifiedNotBefore)
	stmt.WhereIf(in.ModifiedNotAfter > 0, `modified < ?`, in.ModifiedNotAfter)

	if err := stmt.Find(&posts); err != nil {
		panic(err)
	}

	out, err := posts.ToProto(s.setPostExtraFields(ctx, in.ContentOptions))
	if err != nil {
		return nil, err
	}

	if in.WithLink != proto.LinkKind_LinkKindNone {
		for _, p := range out {
			s.setPostLink(p, in.WithLink)
		}
	}

	return &proto.ListPostsResponse{
		Posts: out,
	}, nil
}

func (s *Service) ListAllPostsIds(ctx context.Context) ([]int32, error) {
	ac := auth.Context(ctx)
	var posts models.Posts
	if err := s.tdb.Model(models.Post{}).Select(`id`).
		WhereIf(ac.User.IsGuest(), `status='public'`).Find(&posts); err != nil {
		return nil, err
	}
	var ids []int32
	for _, p := range posts {
		ids = append(ids, int32(p.ID))
	}
	return ids, nil
}

// TODO 性能很差！
func (s *Service) isPostPublic(ctx context.Context, id int64) bool {
	p, err := s.GetPost(ctx, &proto.GetPostRequest{Id: int32(id)})
	if err != nil {
		return false
	}
	return p.Status == `public`
}

// 获取指定编号的文章。
//
// NOTE：如果是公开文章但是非管理员用户，会过滤掉敏感字段。
func (s *Service) GetPost(ctx context.Context, in *proto.GetPostRequest) (*proto.Post, error) {
	ac := auth.Context(ctx)

	var p models.Post

	stmt := s.tdb.Model(p)

	if in.Id > 0 {
		stmt = stmt.Where(`id=?`, in.Id)
	} else if in.Page != "" {
		dir, slug := path.Split(in.Page)
		catID := s.getPageParentID(dir)
		stmt = stmt.Where("slug=? AND category=?", slug, catID).
			OrderBy("date DESC") // ???
	} else {
		return nil, status.Error(codes.InvalidArgument, `需要指定文章查询条件。`)
	}

	stmt = stmt.WhereIf(ac.User.IsGuest(), `status='public'`)
	if err := stmt.Find(&p); err != nil {
		return nil, err
	}

	out, err := p.ToProto(s.setPostExtraFields(ctx, in.ContentOptions))
	if err != nil {
		return nil, err
	}

	if in.WithLink != proto.LinkKind_LinkKindNone {
		s.setPostLink(out, in.WithLink)
	}

	if in.WithRelates {
		relates, err := s.getRelatedPosts(int64(in.Id))
		if err != nil {
			return nil, err
		}
		for _, r := range relates {
			out.Relates = append(out.Relates, &proto.Post{
				Id:    r.ID,
				Title: r.Title,
			})
		}
		if in.WithLink != proto.LinkKind_LinkKindNone {
			for _, p := range out.Relates {
				s.setPostLink(p, in.WithLink)
			}
		}
	}

	return out, nil
}

func (s *Service) setPostLink(p *proto.Post, k proto.LinkKind) {
	switch k {
	case proto.LinkKind_LinkKindRooted:
		p.Link = s.GetPlainLink(p.Id)
	case proto.LinkKind_LinkKindFull:
		p.Link = s.home.JoinPath(s.GetPlainLink(p.Id)).String()
	default:
		panic(`unknown link kind`)
	}
}

func (s *Service) PathResolver(id int64) renderers.PathResolver {
	return &PathResolver{
		base: s.cfg.Data.File.Path,
		fs:   storage.NewLocal(s.cfg.Data.File.Path, fmt.Sprint(id)),
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

func (s *Service) GetPostContentCached(ctx context.Context, id int64, co *proto.PostContentOptions) (string, error) {
	return s.getPostContentCached(ctx, id, co)
}

func (s *Service) getPostContentCached(ctx context.Context, id int64, co *proto.PostContentOptions) (string, error) {
	key := _PostContentCacheKey{
		ID:      id,
		Options: co.String(),
	}
	content, err, _ := s.postContentCaches.GetOrLoad(ctx, key,
		func(ctx context.Context, key _PostContentCacheKey) (string, time.Duration, error) {
			var p models.Post
			s.posts().Where("id = ?", id).MustFind(&p)
			content, err := s.getPostContent(id, p.SourceType, p.Source, p.Metas, co)
			if err != nil {
				return ``, 0, err
			}
			s.postCaches.Append(id, key)
			return content, time.Hour, nil
		},
	)
	if err != nil {
		return ``, err
	}
	return content, nil
}

func (s *Service) getPostTagsCached(ctx context.Context, id int64) ([]string, error) {
	key := fmt.Sprintf(`post_tags:%d`, id)
	tags, err, _ := s.cache.GetOrLoad(ctx, key, func(ctx context.Context, _ string) (any, time.Duration, error) {
		tags := s.GetObjectTagNames(id)
		return tags, time.Hour, nil
	})
	return tags.([]string), err
}

func (s *Service) deletePostContentCacheFor(id int64) {
	log.Println(`即将删除文章缓存：`, id)
	s.postCaches.Delete(id, func(second _PostContentCacheKey) {
		s.postContentCaches.Delete(second)
		log.Println(`删除文章缓存：`, second)
	})
}

func (s *Service) getPostContent(id int64, sourceType, source string, metas models.PostMeta, co *proto.PostContentOptions) (string, error) {
	if !co.WithContent {
		return ``, errors.New(`without content but get content`)
	}

	var tr renderers.Renderer
	switch sourceType {
	case `markdown`:
		options := []renderers.Option2{
			renderers.WithRemoveTitleHeading(true),
			renderers.WithAssetSources(func(path string) (name string, url string, description string, found bool) {
				if src, ok := metas.Sources[path]; ok {
					name = src.Name
					url = src.URL
					description = src.Description
					found = true
				}
				return
			}),
			renderers.WithOpenLinksInNewTab(renderers.OpenLinksInNewTabKind(co.OpenLinksInNewTab)),
		}
		if id > 0 {
			options = append(options, renderers.WithPathResolver(s.PathResolver(id)))
			if link := s.GetLink(id); link != s.plainLink(id) {
				options = append(options, renderers.WithModifiedAnchorReference(link))
			}
			if co.UseAbsolutePaths {
				options = append(options, renderers.WithUseAbsolutePaths(s.plainLink(id)))
			}
		}
		if co.RenderCodeBlocks {
			options = append(options, renderers.WithRenderCodeAsHTML())
		}
		if co.PrettifyHtml {
			options = append(options, renderers.WithHtmlPrettifier())
		}
		options = append(options, s.markdownWithPlantUMLRenderer())

		var fsForTags fs.FS
		if co.UseAbsolutePaths {
			fsForTags = s.fileSystemForRooted()
		} else {
			fsForTags = utils.Must(s.FileSystemForPost(auth.SystemAdmin(context.Background()), id))
		}
		options = append(options, media_tags.New(
			fsForTags,
			s.mediaTagsTemplate,
		))

		tr = renderers.NewMarkdown(options...)
	case `html`:
		tr = &renderers.HTML{}
	default:
		return ``, fmt.Errorf(`unknown source type`)
	}
	return tr.Render(source)
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

func (s *Service) IncrementPostPageView(id int64) {
	query := "UPDATE posts SET page_view=page_view+1 WHERE id=?"
	s.tdb.MustExec(query, id)
}

// GetPostsByTags gets tag posts.
func (s *Service) GetPostsByTags(ctx context.Context, req *proto.GetPostsByTagsRequest) (*proto.GetPostsByTagsResponse, error) {
	ac := auth.Context(ctx)
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
		WhereIf(ac.User.IsGuest(), `posts.status='public'`).
		GroupBy(`posts.id`).
		Having(fmt.Sprintf(`COUNT(posts.id) >= %d`, len(ids))).
		MustFind(&posts)
	outs, err := posts.ToProto(s.setPostExtraFields(ctx, req.ContentOptions))
	if err != nil {
		return nil, err
	}
	if req.WithLink != proto.LinkKind_LinkKindNone {
		for _, p := range outs {
			s.setPostLink(p, req.WithLink)
		}
	}
	return &proto.GetPostsByTagsResponse{Posts: outs}, nil
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

// TODO cache
func (s *Service) getRelatedPosts(id int64) (models.Posts, error) {
	tagIDs := s.getObjectTagIDs(id, true)
	if len(tagIDs) == 0 {
		return nil, nil
	}
	type _PostForRelated struct {
		ID        int64  `json:"id"`
		Title     string `json:"title"`
		Relevance uint   `json:"relevance"`
	}

	var relates []_PostForRelated
	s.tdb.From(models.Post{}).From(models.ObjectTag{}).
		Select("posts.id,posts.title,COUNT(posts.id) relevance").
		Where("post_tags.post_id != ?", id).
		Where("posts.id = post_tags.post_id").
		Where("post_tags.tag_id IN (?)", tagIDs).
		GroupBy("posts.id").
		OrderBy("relevance DESC").
		Limit(9).
		MustFind(&relates)
	var posts models.Posts
	for _, r := range relates {
		posts = append(posts, &models.Post{
			ID:    r.ID,
			Title: r.Title,
		})
	}
	return posts, nil
}

// t: last_commented_at 表示文章评论最后被操作的时间。不是最后被评论的时间。
// 因为属于是外部关联资源，对 304 有贡献。
func (s *Service) updatePostCommentCount(pid int64, t time.Time) {
	var count uint
	s.tdb.Model(models.Comment{}).Where(`post_id=?`, pid).MustCount(&count)
	s.tdb.MustExec(`UPDATE posts SET comments=?,last_commented_at=? WHERE id=?`, count, t.Unix(), pid)
}

// CreatePost ...
func (s *Service) CreatePost(ctx context.Context, in *proto.Post) (*proto.Post, error) {
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

	title, err := s.parsePostTitle(in.SourceType, in.Source)
	if err != nil {
		return nil, err
	}
	if title != `` {
		// 文章中的一级标题优先级大于参数。
		p.Title = title
	}
	// 除碎碎念外，文章不允许空标题
	if p.Type != `tweet` && p.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "文章必须要有标题。")
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
	return p.ToProto(s.setPostExtraFields(ctx, nil))
}

// UpdatePost ...
func (s *Service) UpdatePost(ctx context.Context, in *proto.UpdatePostRequest) (*proto.Post, error) {
	s.MustBeAdmin(ctx)

	if in.Post == nil || in.Post.Id == 0 || in.UpdateMask == nil {
		return nil, status.Error(codes.InvalidArgument, "无效文章编号、更新字段")
	}

	p := models.Post{
		ID: in.Post.Id,
	}

	now := time.Now().Unix()

	m := map[string]any{
		`modified`: now,
	}

	var hasSourceType, hasSource bool
	var hasTags, hasMetas bool
	var hasTitle bool

	for _, path := range in.UpdateMask.Paths {
		switch path {
		case `title`:
			m[path] = in.Post.Title
			hasTitle = true
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
		title, err := s.parsePostTitle(in.Post.SourceType, in.Post.Source)
		if err != nil {
			return nil, err
		}
		// 有些旧文章 MD 内并没有写标题，标题在 config 里面，此处不能强制替换。
		if title != `` {
			// 文章中的一级标题优先级大于配置文件。
			m[`title`] = title
		}
	}
	if hasTitle || (hasSource && hasSourceType) {
		var ty string
		if t, ok := m[`type`].(string); ok {
			ty = t
		} else {
			var p models.Post
			if err := s.tdb.Select(`type`).Where(`id=?`, in.Post.Id).Find(&p); err != nil {
				return nil, err
			}
			ty = p.Type
		}
		title, err := s.parsePostTitle(in.Post.SourceType, in.Post.Source)
		if err != nil {
			return nil, err
		}
		if title != `` {
			// 文章中的一级标题优先级大于参数。
			m[`title`] = title
		}
		// 除碎碎念外，文章不允许空标题
		if ty != `tweet` && title == "" {
			return nil, status.Error(codes.InvalidArgument, "文章必须要有标题。")
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
			s.cache.Delete(fmt.Sprintf(`post_tags:%d`, in.Post.Id))
		}
		txs.updateLastPostTime(time.Now())
		txs.deletePostContentCacheFor(p.ID)
		return nil
	})

	// TODO Update 接口没有 ContentOptions。
	return s.GetPost(ctx, &proto.GetPostRequest{Id: int32(in.Post.Id)})
}

// 只是用来在创建文章和更新文章的时候从正文里面提取。
func (s *Service) parsePostTitle(sourceType, source string) (string, error) {
	var tr renderers.Renderer
	var title string
	switch sourceType {
	case "html":
		tr = &renderers.HTML{}
	case "markdown":
		// 这里只是用 title 的话，可以不考虑 Markdown 的初始化参数。
		tr = renderers.NewMarkdown(
			renderers.WithoutRendering(),
			renderers.WithTitle(&title),
		)
	default:
		return "", status.Error(codes.InvalidArgument, "no renderer was found")
	}
	_, err := tr.Render(source)
	if err != nil {
		return "", status.Error(codes.InvalidArgument, err.Error())
	}
	return title, nil
}

// 用于删除一篇文章。
// 这个函数基本没怎么测试过，因为基本上只是设置为不公开。
func (s *Service) DeletePost(ctx context.Context, in *proto.DeletePostRequest) (*empty.Empty, error) {
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

// TODO 文章编号可能是 0️⃣
func (s *Service) PreviewPost(ctx context.Context, in *proto.PreviewPostRequest) (*proto.PreviewPostResponse, error) {
	s.MustBeAdmin(ctx)
	// ac := auth.Context(ctx)
	content, err := s.getPostContent(int64(in.Id), `markdown`, in.Markdown, models.PostMeta{}, &proto.PostContentOptions{
		WithContent:       true,
		RenderCodeBlocks:  true,
		OpenLinksInNewTab: proto.PostContentOptions_OpenLinkInNewTabKindAll,
		UseAbsolutePaths:  true,
	})
	return &proto.PreviewPostResponse{Html: content}, err
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
func (s *Service) SetPostStatus(ctx context.Context, in *proto.SetPostStatusRequest) (*proto.SetPostStatusResponse, error) {
	s.MustBeAdmin(ctx)

	s.MustTxCall(func(txs *Service) error {
		var post models.Post
		txs.tdb.Select("id").Where("id=?", in.Id).MustFind(&post)

		status := `public`
		if !in.Public {
			status = `draft`
		}

		m := map[string]any{
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
	return &proto.SetPostStatusResponse{}, nil
}

// GetPostCommentsCount ...
func (s *Service) GetPostCommentsCount(ctx context.Context, in *proto.GetPostCommentsCountRequest) (*proto.GetPostCommentsCountResponse, error) {
	var post models.Post
	s.posts().Select("comments").Where("id=?", in.PostId).MustFind(&post)
	return &proto.GetPostCommentsCountResponse{
		Count: int64(post.Comments),
	}, nil
}

// 由于“相关文章”目前只在 GetPost 时返回，所以不在这里设置。
func (s *Service) setPostExtraFields(ctx context.Context, co *proto.PostContentOptions) func(c *proto.Post) error {
	ac := auth.Context(ctx)

	return func(p *proto.Post) error {
		if !ac.User.IsAdmin() && !ac.User.IsSystem() {
			p.Metas.Geo = nil
		}

		if co != nil && co.WithContent {
			content, err := s.getPostContentCached(ctx, p.Id, co)
			if err != nil {
				return err
			}
			p.Content = content
		}
		// 碎碎念可能没有标题，自动生成
		if p.Type == `tweet` {
			switch p.Title {
			case ``, `Untitled`, models.Untitled:
				content, err := s.getPostContentCached(ctx, p.Id, &proto.PostContentOptions{
					WithContent:  true,
					PrettifyHtml: true,
				})
				if err != nil {
					return err
				}
				p.Title = truncateTitle(content, 36)
			}
		}

		if tags, err := s.getPostTagsCached(ctx, p.Id); err != nil {
			return err
		} else {
			p.Tags = tags
		}

		return nil
	}
}

// TODO：可能把 [图片] 这种截断
func truncateTitle(title string, length int) string {
	runes := []rune(title)

	// 不包含回车
	if p := slices.Index(runes, '\n'); p > 0 { // 不可能第一个字符是回车吧？🤔
		runes = runes[:p]
	}

	// 不超过指定的字符串长度
	maxLength := utils.IIF(length > len(runes), len(runes), length)

	// 不包含句号
	if p := slices.Index(runes, '。'); p > 0 && p < maxLength {
		maxLength = p
	}

	suffix := utils.IIF(len(runes) > maxLength, "...", "")
	return string(runes[:maxLength]) + suffix
}
