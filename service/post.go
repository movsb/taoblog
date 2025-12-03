package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/db"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/assets"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/hashtags"
	"github.com/movsb/taoblog/service/modules/renderers/page_link"
	"github.com/movsb/taoblog/service/modules/renderers/toc"
	"github.com/movsb/taoblog/theme/styling"
	"github.com/movsb/taorm"
	"github.com/phuslu/lru"
	"github.com/sergi/go-diff/diffmatchpatch"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type _PostContentCacheKey struct {
	ID      int64
	Options string

	// 除开最基本的文章编号和渲染选项的不同之外，
	// 可能还有其它的 Vary 特性，比如：如果同一篇文章，管理员和访客看到的内容不一样（角色），
	// 这部分就属于 Vary 应该标记出来的。暂时不使用相关标记。只是备用。
	Vary struct{}

	// NOTE 由于增加了用户系统，不同用于看不同用户的文章时应该有不同的缓存。
	// 见隔离测试：TestIsolatedPostCache
	UserID int

	// 公开与否时渲染不同，比如加密选项。
	// NOTE：评论的此状态=文章状态，因为评论目前没有自己的存储。
	Public bool
}

func (s *Service) posts() *taorm.Stmt {
	return s.tdb.Model(models.Post{})
}

// 按条件枚举文章。
//
// TODO: 具体的 permission 没用上。
// TODO: 好像对于登录用于 status=public 没用上。
// TODO: distinct posts.* 是正确的用法吗？
func (s *Service) ListPosts(ctx context.Context, in *proto.ListPostsRequest) (*proto.ListPostsResponse, error) {
	ac := user.Context(ctx)

	var posts models.Posts

	stmt := s.posts().
		Limit(int64(in.Limit)).
		// ORM 会安全校验 order by 语句是否规范，这里不用校验。
		OrderBy(in.OrderBy)

	if ac.User.IsSystem() {
		// nothing to do
	} else if !ac.User.IsGuest() {
		switch in.Ownership {
		default:
			return nil, fmt.Errorf(`未知所有者。`)
		case proto.Ownership_OwnershipDrafts:
			stmt.Where(`user_id=? AND status=?`, ac.User.ID, models.PostStatusDraft)
		case proto.Ownership_OwnershipMine:
			stmt.Where(`user_id=? AND status!=?`, ac.User.ID, models.PostStatusDraft)
		case proto.Ownership_OwnershipShared:
			stmt.Select(`distinct posts.*`)
			stmt.InnerJoin(models.AccessControlEntry{}, `posts.id = acl.post_id`)
			stmt.Where(
				`posts.user_id!=? AND (acl.user_id=? AND posts.status = ?)`,
				ac.User.ID, ac.User.ID, models.PostStatusPartial,
			)
		case proto.Ownership_OwnershipTheir:
			stmt.Select(`distinct posts.*`)
			stmt.InnerJoin(models.AccessControlEntry{}, `posts.id = acl.post_id`)
			stmt.Where(
				`posts.user_id!=? AND (posts.status=? OR (acl.user_id=? AND posts.status = ?))`,
				ac.User.ID, models.PostStatusPublic, ac.User.ID, models.PostStatusPartial,
			)
		case proto.Ownership_OwnershipMineAndShared:
			stmt.Select(`distinct posts.*`)
			stmt.LeftJoin(models.AccessControlEntry{}, `posts.id = acl.post_id`)
			stmt.Where(
				`(posts.user_id=? AND posts.status!=?) OR (acl.user_id=? AND posts.status = ?)`,
				ac.User.ID, models.PostStatusDraft, ac.User.ID, models.PostStatusPartial,
			)
		case proto.Ownership_OwnershipUnknown, proto.Ownership_OwnershipAll:
			stmt.Select(`distinct posts.*`)
			stmt.LeftJoin(models.AccessControlEntry{}, `posts.id = acl.post_id`)
			stmt.Where(
				`(posts.user_id=? AND posts.status!=?) OR (posts.status=? OR (acl.user_id=? AND posts.status = ?))`,
				ac.User.ID, models.PostStatusDraft, models.PostStatusPublic, ac.User.ID, models.PostStatusPartial,
			)
		}
	} else {
		stmt.Where("status=?", models.PostStatusPublic)
	}

	// TODO 以前觉得这样写很省事儿。但是这样好像无法写覆盖测试？
	stmt.WhereIf(len(in.Kinds) > 0, `type in (?)`, in.Kinds)
	stmt.WhereIf(in.ModifiedNotBefore > 0, `modified >= ?`, in.ModifiedNotBefore)
	stmt.WhereIf(in.ModifiedNotAfter > 0, `modified < ?`, in.ModifiedNotAfter)
	stmt.WhereIf(len(in.Categories) > 0, `category in (?)`, in.Categories)

	if err := stmt.Find(&posts); err != nil {
		panic(err)
	}

	out, err := posts.ToProto(s.setPostExtraFields(ctx, in.GetPostOptions))
	if err != nil {
		return nil, err
	}

	return &proto.ListPostsResponse{Posts: out}, nil
}

// 只会列出公开的。
func (s *Service) ListAllPostsIds(ctx context.Context) ([]int32, error) {
	user.MustBeAdmin(ctx)
	var posts models.Posts
	if err := s.tdb.Select(`id`).Where(`status=?`, models.PostStatusPublic).Find(&posts); err != nil {
		return nil, err
	}
	var ids []int32
	for _, p := range posts {
		ids = append(ids, int32(p.ID))
	}
	return ids, nil
}

// 获取指定编号的文章。
//
// NOTE：如果是访客用户，会过滤掉敏感字段。
func (s *Service) GetPost(ctx context.Context, in *proto.GetPostRequest) (_ *proto.Post, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.Context(ctx)

	var p models.Post

	stmt := s.tdb.Select(`posts.*`)

	if in.Id > 0 {
		stmt.Where(`posts.id=?`, in.Id)
	} else if in.Page != "" {
		dir, slug := path.Split(in.Page)
		catID := s.getPageParentID(dir)
		stmt.Where("slug=? AND category=?", slug, catID).
			OrderBy("date DESC") // ???
	} else {
		return nil, status.Error(codes.InvalidArgument, `需要指定文章查询条件。`)
	}

	if ac.User.IsGuest() {
		stmt.Where(`status=?`, models.PostStatusPublic)
	} else if ac.User.IsSystem() {
		// fallthrough
	} else {
		stmt.LeftJoin(models.AccessControlEntry{}, `posts.id = acl.post_id`)
		stmt.Where(
			`posts.user_id=? OR posts.status=? OR (acl.user_id=? AND acl.permission=? AND posts.status=?)`,
			ac.User.ID, models.PostStatusPublic, ac.User.ID, models.PermRead, models.PostStatusPartial,
		)
	}

	if err := stmt.Find(&p); err != nil {
		if taorm.IsNotFoundError(err) {
			return nil, status.Error(codes.NotFound, `文章未找到`)
		}
		return nil, err
	}

	return p.ToProto(s.setPostExtraFields(ctx, in.GetPostOptions))
}

func (s *Service) setPostLink(p *proto.Post, k proto.LinkKind) {
	switch k {
	case proto.LinkKind_LinkKindRooted:
		p.Link = s.GetPlainLink(p.Id)
	case proto.LinkKind_LinkKindFull:
		p.Link = utils.Must1(url.Parse(s.getHome())).JoinPath(s.GetPlainLink(p.Id)).String()
	default:
		panic(`unknown link kind`)
	}
}

func (s *Service) OpenAsset(id int64) gold_utils.WebFileSystem {
	return gold_utils.NewWebFileSystem(
		_OpenPostFile{s: s},
		utils.Must1(url.Parse(s.getHome())).JoinPath(fmt.Sprintf("/%d/", id)),
	)
}

type _OpenPostFile struct {
	s *Service
}

// 只支持打开 /123/a.txt 这种 URL 对应的文件。
func (f _OpenPostFile) Open(name string) (fs.File, error) {
	before, after, found := strings.Cut(name, `/`)
	if !found {
		return nil, os.ErrNotExist
	}
	id, err := strconv.Atoi(before)
	if err != nil {
		return nil, err
	}
	return f.s.postDataFS.ForPost(id).Open(after)
}

func (s *Service) getPostContentCached(ctx context.Context, p *proto.Post, co *proto.PostContentOptions) (string, error) {
	ac := user.Context(ctx)
	id := p.Id
	key := _PostContentCacheKey{
		ID:      id,
		Options: co.String(),
		UserID:  int(ac.User.ID),
		Public:  p.Status == models.PostStatusPublic,
	}
	content, err, _ := s.postContentCaches.GetOrLoad(ctx, key,
		func(ctx context.Context, key _PostContentCacheKey) (string, time.Duration, error) {
			var p models.Post
			s.posts().Where("id = ?", id).MustFind(&p)
			content, err := s.renderMarkdown(ctx, true, id, 0, p.SourceType, p.Source, p.Metas, co, key.Public)
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

// 用于预览文章的时候快速 diff。
// 编辑操作很频繁，不需要一直刷数据库。
// 只能原作者获取。
func (s *Service) getPostSourceCached(ctx context.Context, id int64) (_ string, outErr error) {
	defer utils.CatchAsError(&outErr)
	type _SourceCacheKey struct {
		UserID int
		Source string
	}
	ac := user.Context(ctx)
	key := fmt.Sprintf(`post_source:%d`, id)
	cache := utils.Must1(utils.DropLast2(s.cache.GetOrLoad(ctx, key, func(ctx context.Context, _ string) (any, time.Duration, error) {
		log.Println(`无 source 缓存，从数据库加载……`)
		db := db.FromContextDefault(ctx, s.tdb)
		var p models.Post
		if err := db.Where(`id=?`, id).Select(`user_id,source`).Find(&p); err != nil {
			return nil, 0, err
		}
		return _SourceCacheKey{UserID: int(p.UserID), Source: p.Source}, time.Minute * 10, nil
	}))).(_SourceCacheKey)
	if cache.UserID != int(ac.User.ID) {
		panic(noPerm)
	}
	return cache.Source, nil
}

func (s *Service) getPostTagsCached(ctx context.Context, id int64) ([]string, error) {
	key := fmt.Sprintf(`post_tags:%d`, id)
	tags, err, _ := s.cache.GetOrLoad(ctx, key, func(ctx context.Context, _ string) (any, time.Duration, error) {
		tags := s.GetObjectTagNames(id)
		return tags, time.Hour, nil
	})
	return tags.([]string), err
}

func (s *Service) getPostTocCached(id int, source string) string {
	key := fmt.Sprintf(`post_toc:%d`, id)
	toc, err, _ := s.cache.GetOrLoad(context.Background(), key, func(ctx context.Context, s string) (any, time.Duration, error) {
		var html []byte
		md := renderers.NewMarkdown(toc.New(&html))
		md.Render(source)
		return html, time.Hour, nil
	})
	if err != nil {
		return ``
	}
	return string(toc.([]byte))
}

func (s *Service) DropAllPostAndCommentCache() {
	s.postCaches.Clear()
	s.commentCaches.Clear()

	// 竟然不能清空？
	s.postContentCaches = lru.NewTTLCache[_PostContentCacheKey, string](10240)
	s.commentContentCaches = lru.NewTTLCache[_PostContentCacheKey, string](10240)

	// log.Println(`已清空所有文章和评论缓存`)
}

func (s *Service) deletePostContentCacheFor(id int64) {
	s.postFullCaches.Delete(id)
	s.postCaches.Delete(id, func(second _PostContentCacheKey) {
		s.postContentCaches.Delete(second)
		log.Println(`删除文章缓存：`, second)
	})
	s.cache.Delete(fmt.Sprintf(`post_source:%d`, id))
	s.cache.Delete(fmt.Sprintf(`post_toc:%d`, id))
	s.cache.Delete(fmt.Sprintf(`post_tags:%d`, id))
}

func withEmojiFilter(node *goquery.Selection) bool {
	return node.HasClass(`emoji`)
}

func (s *Service) hashtagResolver(tag string) string {
	u := utils.Must1(url.Parse(`/tags`))
	return u.JoinPath(tag).String()
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

func (s *Service) IncrementViewCount(m map[int]int) {
	if len(m) <= 0 {
		return
	}
	s.tdb.MustTxCall(func(tx *taorm.DB) {
		for p, n := range m {
			tx.MustExec(fmt.Sprintf(`UPDATE posts SET page_view=page_view+%d WHERE id=?`, n), p)
		}
	})
}

// 只能获取公开的或自己创建的。
func (s *Service) GetPostsByTags(ctx context.Context, tagNames []string) ([]*proto.Post, error) {
	ac := user.Context(ctx)
	var ids []int64
	for _, tag := range tagNames {
		t := s.GetTagByName(tag)
		ids = append(ids, t.ID)
	}
	tagIDs := s.getAliasTagsAll(ids)
	var posts models.Posts
	s.tdb.From(models.Post{}).From(models.ObjectTag{}).Select("posts.*").
		Where("posts.id=post_tags.post_id").
		Where("post_tags.tag_id in (?)", tagIDs).
		WhereIf(ac.User.IsGuest(), `posts.status='public'`).
		WhereIf(!ac.User.IsGuest(), `posts.status=? OR posts.user_id=?`, models.PostStatusPublic, ac.User.ID).
		GroupBy(`posts.id`).
		Having(fmt.Sprintf(`COUNT(posts.id) >= %d`, len(ids))).
		MustFind(&posts)
	return posts.ToProto(s.setPostExtraFields(ctx, &proto.GetPostOptions{}))
}

func (s *Service) getPageParentID(parents string) int64 {
	parents = strings.Trim(parents, `/`)
	if len(parents) == 0 {
		return 0
	}
	slugs := strings.Split(parents, "/")

	type getPageParentID_Result struct {
		ID       int64
		Slug     string
		Category int64
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
			if r.Category == parent && r.Slug == slugs[i] {
				parent = r.ID
				found = true
				break
			}
		}
		if !found {
			panic(status.Errorf(codes.NotFound, "找不到父页面：%s", slugs[i]))
		}
	}

	return parent
}

// TODO cache
// TODO 添加权限测试
// 只能获取公开的或自己创建的。
func (s *Service) getRelatedPostsCached(ctx context.Context, id int64) ([]*proto.Post, error) {
	return utils.DropLast2(s.relatesCaches.Load().GetOrLoad(ctx, id, func(context.Context, int64) ([]*proto.Post, time.Duration, error) {
		posts, err := s.getRelatedPosts(ctx, id)
		return posts, time.Hour, err
	}))
}

func (s *Service) getRelatedPosts(ctx context.Context, id int64) ([]*proto.Post, error) {
	ac := user.Context(ctx)

	tagIDs := s.getObjectTagIDs(id, true)
	if len(tagIDs) == 0 {
		return nil, nil
	}
	type _PostForRelated struct {
		models.Post
		Relevance uint `json:"relevance"`
	}

	var relates []_PostForRelated
	s.tdb.From(models.Post{}).From(models.ObjectTag{}).
		Select("posts.*,COUNT(posts.id) relevance").
		Where("post_tags.post_id != ?", id).
		Where("posts.id = post_tags.post_id").
		Where("post_tags.tag_id IN (?)", tagIDs).
		WhereIf(ac.User.IsGuest(), `posts.status=?`, models.PostStatusPublic).
		WhereIf(!ac.User.IsGuest(), `posts.status=? OR posts.user_id=?`, models.PostStatusPublic, ac.User.ID).
		GroupBy("posts.id").
		OrderBy("relevance DESC").
		Limit(9).
		MustFind(&relates)
	var posts models.Posts
	for _, r := range relates {
		posts = append(posts, &r.Post)
	}
	return posts.ToProto(s.setPostExtraFields(ctx, nil))
}

// t: last_commented_at 表示文章评论最后被操作的时间。不是最后被评论的时间。
// 因为属于是外部关联资源，对 304 有贡献。
func (s *Service) updatePostCommentCount(pid int64, t time.Time) {
	var count uint
	s.tdb.Model(models.Comment{}).Where(`post_id=?`, pid).MustCount(&count)
	s.tdb.MustExec(`UPDATE posts SET comments=?,last_commented_at=? WHERE id=?`, count, t.Unix(), pid)
}

// 有些特别的代码会贡献 304，比如图片元数据，此时需要更新文章。
func (s *Service) updatePostMetadataTime(pid int64, t time.Time) {
	s.tdb.MustExec(`UPDATE posts SET last_commented_at=? WHERE id=?`, t.Unix(), pid)
}

func (s *Service) CreatePost(ctx context.Context, in *proto.Post) (*proto.Post, error) {
	ac := user.MustNotBeGuest(ctx)

	now := int32(time.Now().Unix())

	p := models.Post{
		ID:         in.Id,
		UserID:     int32(ac.User.ID),
		Date:       0,
		Modified:   0,
		Title:      strings.TrimSpace(in.Title),
		Slug:       in.Slug,
		Type:       in.Type,
		Category:   in.Category,
		Status:     models.PostStatusDraft,
		Metas:      *models.PostMetaFrom(in.Metas),
		Source:     in.Source,
		SourceType: in.SourceType,
	}

	if strings.TrimSpace(p.Source) == "" {
		return nil, status.Error(codes.InvalidArgument, "内容不应为空。")
	}

	p.ModifiedTimezone = in.ModifiedTimezone
	if in.Modified > 0 {
		p.Modified = in.Modified
	}

	p.DateTimezone = in.DateTimezone
	if in.Date > 0 {
		p.Date = in.Date
		if in.Modified == 0 {
			p.Modified = p.Date
			p.ModifiedTimezone = p.DateTimezone
		}
	} else {
		p.Date = now
		p.Modified = now
		// TODO 设置时区。
		p.DateTimezone = ``
		p.ModifiedTimezone = ``
	}

	if in.Status != "" {
		p.Status = in.Status
	}

	if p.Type == `` {
		p.Type = `post`
	}
	if p.Type == `page` && p.Slug == `` {
		return nil, status.Error(codes.InvalidArgument, `页面必须要有路径名（slug）。`)
	}

	if p.SourceType == `` {
		in.SourceType = `markdown`
		p.SourceType = in.SourceType
	}

	if err := s.checkPostCat(ctx, p.Category); err != nil {
		return nil, fmt.Errorf(`创建文章失败：%w`, err)
	}

	derived, err := s.parseDerived(ctx, in.SourceType, in.Source)
	if err != nil {
		return nil, err
	}
	if derived.Title != `` {
		// 文章中的一级标题优先级大于参数。
		p.Title = derived.Title
	}
	// 除碎碎念外，文章不允许空标题
	if p.Type != `tweet` && p.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "文章必须要有标题。")
	}

	s.MustTxCall(func(txs *Service) error {
		txs.tdb.Model(&p).MustCreate()
		in.Id = p.ID
		txs.updateObjectTags(p.ID, append(derived.Tags, in.Tags...))
		txs.updateLastPostTime(time.Unix(int64(p.Modified), 0))
		txs.updatePostPageCount()
		txs.updateReferences(ctx, int32(p.ID), nil, derived.References)
		return nil
	})

	s.updateUserTopPosts(int(ac.User.ID), int(p.ID), in.Top)

	// TODO 暂时没提供选项。
	return p.ToProto(s.setPostExtraFields(ctx, nil))
}

// 本身未鉴权，由 CreatePost 鉴权。
func (s *Service) CreateUntitledPost(ctx context.Context, in *proto.CreateUntitledPostRequest) (*proto.CreateUntitledPostResponse, error) {
	var source string
	switch in.Type {
	case `markdown`:
		source = models.UntitledSourceMarkdown
	default:
		return nil, status.Error(codes.InvalidArgument, `unknown post type`)
	}
	p, err := s.CreatePost(ctx, &proto.Post{
		Type:       `post`,
		SourceType: in.Type,
		Source:     source,
	})
	return &proto.CreateUntitledPostResponse{
		Post: p,
	}, err
}

// 缓存的是数据库中的完整原始文章数据。
func (s *Service) getPostCached(ctx context.Context, id int) (*models.Post, error) {
	p, err, _ := s.postFullCaches.GetOrLoad(ctx, int64(id), func(ctx context.Context, i int64) (*models.Post, time.Duration, error) {
		var post models.Post
		return &post, time.Hour, s.tdb.Where(`id=?`, id).Find(&post)
	})
	return p, err
}

// TODO 上缓存。
func (s *Service) getPostTitle(ctx context.Context, id int32) (string, error) {
	post, err := s.GetPost(ctx, &proto.GetPostRequest{Id: int32(id)})
	if err != nil {
		return ``, fmt.Errorf(`getPostTitle: %w`, err)
	}
	return post.Title, nil
}

func isUpdatingUntitledPost(p *models.Post) bool {
	return p.Date == p.Modified &&
		p.Title == models.Untitled &&
		(p.Source == models.UntitledSourceMarkdown) &&
		(p.Metas.Geo == nil || p.Metas.Geo.Latitude == 0)
}

// 更新文章。
// 需要携带版本号，像评论一样。
func (s *Service) UpdatePost(ctx context.Context, in *proto.UpdatePostRequest) (*proto.Post, error) {
	ac := user.MustNotBeGuest(ctx)

	if in.Post == nil || in.Post.Id == 0 || in.UpdateMask == nil {
		return nil, status.Error(codes.InvalidArgument, "无效文章编号、更新字段")
	}

	// TODO：放事务中。
	oldPost, err := s.getPostCached(ctx, int(in.Post.Id))
	if err != nil {
		return nil, err
	}

	// 仅可编辑自己的文章。
	if !(ac.User.IsSystem() || ac.User.ID == int64(oldPost.UserID)) {
		panic(status.Error(codes.PermissionDenied, noPerm))
	}

	now := time.Now().Unix()

	m := map[string]any{}

	// 适用于导入三方数据的时候更新导入。
	if !in.DoNotTouch {
		m[`modified`] = now
		// TODO 使用 now 的时区对应名修改 modified_timezone
	}

	var hasSourceType, hasSource bool
	var hasMetas bool
	var hasTitle bool
	var hasType bool
	var hasSlug bool

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
			hasSlug = true
		case `metas`:
			hasMetas = true
		case `type`:
			m[path] = in.Post.Type
			hasType = true
		case `date`:
			m[`date`] = in.Post.Date
		case `date_timezone`:
			m[path] = in.Post.DateTimezone
		case `modified_timezone`:
			m[path] = in.Post.ModifiedTimezone
		case `status`:
			m[`status`] = in.Post.Status
		default:
			panic(`unknown update mask:` + path)
		}
	}

	// 更新无标题文章时如果时间未指定，则更新为现在时间。
	if isUpdatingUntitledPost(oldPost) && in.Post.Date == oldPost.Date {
		m[`date`] = now
		// 可能没有修改时区，但是空也是合法的。
		if tz, ok := m[`modified_timezone`]; ok {
			m[`date_timezone`] = tz
		}
	}

	if hasMetas {
		m[`metas`] = models.PostMetaFrom(in.Post.Metas)
	}

	if hasSourceType != hasSource {
		return nil, status.Error(codes.InvalidArgument, `source type and source must be specified`)
	}

	derived, err := s.parseDerived(ctx, in.Post.SourceType, in.Post.Source)
	if err != nil {
		return nil, err
	}

	if hasSource && hasSourceType {
		if derived.Title != `` {
			// 文章中的一级标题优先级大于配置文件。
			m[`title`] = derived.Title
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
		if derived.Title != `` {
			// 文章中的一级标题优先级大于参数。
			m[`title`] = derived.Title
		} else {
			if !hasTitle {
				m[`title`] = ``
			}
		}
		// 除碎碎念外，文章不允许空标题
		if ty != `tweet` && (derived.Title == "" && !hasTitle) {
			return nil, status.Error(codes.InvalidArgument, "文章必须要有标题。")
		}
	}
	if hasType && in.Post.Type == `page` && (hasSlug && in.Post.Slug == `` || oldPost.Slug == ``) {
		return nil, status.Error(codes.InvalidArgument, `页面必须要有路径名（slug）。`)
	}

	if in.UpdateCategory {
		if err := s.checkPostCat(ctx, in.Post.Category); err != nil {
			return nil, fmt.Errorf(`更新文章失败：%w`, err)
		}
		m[`category`] = in.Post.Category
	}

	s.MustTxCall(func(txs *Service) error {
		p := models.Post{ID: in.Post.Id}
		res := txs.tdb.Model(p).Where(`modified=?`, in.Post.Modified).MustUpdateMap(m)
		rowsAffected, err := res.RowsAffected()
		if err != nil || rowsAffected != 1 {
			op := models.Post{ID: in.Post.Id}
			txs.tdb.Model(&op).MustFind(&op)
			return status.Errorf(codes.Aborted, "update failed, modified conflict: %v (modified: %v)", err, op.Modified)
		}

		txs.updateObjectTags(p.ID, derived.Tags)
		txs.updateReferences(ctx, int32(p.ID), &oldPost.Citations, derived.References)
		txs.updateLastPostTime(time.Now())

		txs.deletePostContentCacheFor(p.ID)

		return nil
	})

	// TODO TODO TODO 事务冲突，暂时放外面！！
	if in.UpdateUserPerms {
		s.setPostACL(in.Post.Id, in.UserPerms, in.SendUserNotify)
	}

	if in.UpdateTop {
		s.updateUserTopPosts(int(ac.User.ID), int(in.Post.Id), in.Post.Top)
	}

	// 文件更新后“相关文章”也会变化，但是难以计算出哪些文章被影响。
	// 为简单起见，直接清空所有“相关文章”缓存。
	s.relatesCaches.Store(lru.NewTTLCache[int64, []*proto.Post](128))

	// 通知新文章创建
	// TODO 异步执行。
	if isUpdatingUntitledPost(oldPost) && oldPost.UserID != int32(user.AdminID) {
		title, _ := m[`title`].(string)
		s.notifier.SendInstant(user.SystemForLocal(ctx), &proto.SendInstantRequest{
			Title: `新文章发表`,
			Body:  fmt.Sprintf(`%s 发表了新文章 %s`, ac.User.Nickname, title),
		})
	}

	// SetPostACL 也会修改文章时间，这里能确保拿到的是最新的。
	return s.GetPost(ctx, &proto.GetPostRequest{
		Id:             int32(in.Post.Id),
		GetPostOptions: in.GetPostOptions,
	})
}

func (s *Service) setPostACL(postID int64, users []int32, notify bool) {
	m := map[int32]*proto.UserPerm{}
	for _, id := range users {
		m[id] = &proto.UserPerm{
			Perms: []proto.Perm{proto.Perm_PermRead},
		}
	}
	utils.Must1(s.SetPostACL(user.SystemForLocal(context.Background()),
		&proto.SetPostACLRequest{
			PostId:         postID,
			Users:          m,
			SendUserNotify: notify,
		},
	))
}

// 只是用来在创建文章和更新文章的时候从正文里面提取。
type _Derived struct {
	Title      string   // # 标题
	Tags       []string // #标签 #标签
	References []int32  // [[页面引用]]
}

func (s *Service) parseDerived(ctx context.Context, sourceType, source string) (*_Derived, error) {
	var title string
	var tags []string
	var refs []int32
	switch sourceType {
	case "markdown":
		tr := renderers.NewMarkdown(
			renderers.WithoutRendering(),
			renderers.WithTitle(&title),
			hashtags.New(s.hashtagResolver, &tags),
			// 这里的 ctx 会用来给 getPostTitle 鉴权用，所以必须是原始请求附带的 ctx。
			page_link.New(ctx, s.getPostTitle, &refs),
		)
		_, err := tr.Render(source)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		d := &_Derived{
			Title:      title,
			Tags:       tags,
			References: refs,
		}
		return d, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "no renderer was found for: %q", sourceType)
	}
}

// 用于删除一篇文章。
// 这个函数基本没怎么测试过，因为基本上只是设置为不公开。
func (s *Service) DeletePost(ctx context.Context, in *proto.DeletePostRequest) (*empty.Empty, error) {
	user.MustBeAdmin(ctx)

	var p models.Post

	s.MustTxCall(func(txs *Service) error {
		txs.tdb.Where(`id=?`, in.Id).MustFind(&p)
		txs.tdb.Model(&p).MustDelete()
		txs.deletePostComments(ctx, int64(in.Id))
		txs.deletePostTags(ctx, int64(in.Id))
		txs.deletePostContentCacheFor(int64(in.Id))
		txs.updatePostPageCount()
		txs.updateCommentsCount()
		txs.deleteReferences(ctx, int32(p.ID), &p.Citations)
		return nil
	})

	s.updateUserTopPosts(int(p.UserID), int(p.ID), false)

	return &empty.Empty{}, nil
}

// TODO 文章编号可能是 0️⃣
func (s *Service) PreviewPost(ctx context.Context, in *proto.PreviewPostRequest) (*proto.PreviewPostResponse, error) {
	user.MustNotBeGuest(ctx)

	out := proto.PreviewPostResponse{}

	ctx = assets.With(ctx, &out.Paths, fmt.Sprintf(`/%d/`, in.Id))

	content, err := s.renderMarkdown(ctx, true, int64(in.Id), 0, in.Type, in.Source, models.PostMeta{}, co.For(co.PreviewPost), s.isPostPublic(ctx, int(in.Id)))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	out.Html = content

	source, err := s.getPostSourceCached(ctx, int64(in.Id))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 不太清楚这个 check lines 参数是干啥的。
	diffs := diffmatchpatch.New().DiffMain(source, in.Source, true)
	var buf bytes.Buffer
	for _, diff := range diffs {
		text := html.EscapeString(diff.Text)
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			fmt.Fprint(&buf, `<ins>`, text, `</ins>`)
		case diffmatchpatch.DiffDelete:
			fmt.Fprint(&buf, `<del>`, text, `</del>`)
		case diffmatchpatch.DiffEqual:
			fmt.Fprint(&buf, text)
		}
	}
	out.Diff = buf.String()

	// 自动保存。
	if in.Save {
		rsp, err := s.UpdatePost(ctx, &proto.UpdatePostRequest{
			Post: &proto.Post{
				Id:         int64(in.Id),
				SourceType: in.Type,
				Source:     in.Source,
				Modified:   in.ModifiedAt,
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					`source_type`,
					`source`,
				},
			},
		})
		if err != nil {
			return &out, err
		}
		out.Title = rsp.Title
		out.UpdatedAt = rsp.Modified
	}

	return &out, nil
}

// updateLastPostTime updates last_post_time in options.
func (s *Service) updateLastPostTime(t time.Time) {
	s.options.SetInteger("last_post_time", t.Unix())
}

func (s *Service) updatePostPageCount() {
	var postCount, pageCount int
	s.tdb.Model(models.Post{}).Select(`count(1) as count`).Where(`type='post'`).MustFind(&postCount)
	s.tdb.Model(models.Post{}).Select(`count(1) as count`).Where(`type='page'`).MustFind(&pageCount)
	s.options.SetInteger(`post_count`, int64(postCount))
	s.options.SetInteger(`page_count`, int64(pageCount))
}

// 更新文章的对外引用信息。
//
// NOTE: 无须判断权限。无权限的文章不会显示任何信息。
//
//   - self: 当前文章编号。
//   - refs：旧的引用/被引用信息。
//   - new： 当前文章最新的对外引用信息。
func (s *Service) updateReferences(ctx context.Context, self int32, refs *models.References, new []int32) {
	posts := map[int32]*models.Post{}

	// 不存在返回空。
	getPost := func(pid int32) *models.Post {
		p, ok := posts[pid]
		if ok {
			return p
		}
		p, err := s.getPostCached(ctx, int(pid))
		if err != nil {
			if taorm.IsNotFoundError(err) {
				return nil
			}
			panic(err)
		}
		posts[pid] = p
		return p
	}

	var oldTo []int32
	if refs != nil && refs.Posts != nil {
		oldTo = refs.Posts.To
	}

	removed := utils.Filter(oldTo, func(n int32) bool { return !slices.Contains(new, n) })
	added := utils.Filter(new, func(n int32) bool { return !slices.Contains(oldTo, n) })

	if len(removed)+len(added) == 0 {
		return
	}

	now := time.Now().Unix()

	{
		if refs == nil {
			refs = &models.References{}
		}
		if refs.Posts == nil {
			refs.Posts = &proto.Post_References_Posts{}
		}
		refs.Posts.To = new

		s.tdb.Model(&models.Post{ID: int64(self)}).MustUpdateMap(taorm.M{
			`last_commented_at`: now,
			`citations`:         refs,
		})

		s.postFullCaches.Delete(int64(self))
	}

	for _, pid := range removed {
		p := getPost(pid)
		if p == nil {
			// NOTE: 可以 Panic。
			continue
		}
		if p.Citations.Posts == nil {
			continue
		}
		from := &p.Citations.Posts.From
		*from = slices.DeleteFunc(*from, func(n int32) bool { return n == self })
		log.Printf(`删除引用：%d → %d`, self, pid)
	}

	for _, pid := range added {
		p := getPost(pid)
		if p == nil {
			// NOTE: 可以 Panic。
			continue
		}
		// 新添加的时候可能为nil，需要判断。
		if p := &p.Citations.Posts; *p == nil {
			*p = &proto.Post_References_Posts{}
		}
		from := &p.Citations.Posts.From
		*from = append(*from, self)
		log.Printf(`增加引用：%d → %d`, self, pid)
	}

	for _, p := range posts {
		s.tdb.Model(p).MustUpdateMap(taorm.M{
			`last_commented_at`: now,
			`citations`:         &p.Citations,
		})
	}

	// 文章标题可能有更改，丢弃引用本文章的文章的缓存。
	for _, from := range refs.Posts.From {
		s.deletePostContentCacheFor(int64(from))
	}
}

// 删除本文章的引用/被引用信息。
func (s *Service) deleteReferences(ctx context.Context, self int32, refs *models.References) {
	now := time.Now().Unix()

	if refs.Posts == nil {
		return
	}

	for _, ref := range refs.Posts.To {
		p, err := s.getPostCached(ctx, int(ref))
		if err != nil {
			if taorm.IsNotFoundError(err) {
				continue
			}
			panic(err)
		}
		if p.Citations.Posts == nil {
			panic(`不应该为 nil`)
		}
		from := &p.Citations.Posts.From
		*from = slices.DeleteFunc(*from, func(n int32) bool { return n == self })
		log.Printf(`删除引用：%d → %d`, self, ref)

		s.tdb.Model(p).MustUpdateMap(taorm.M{
			`last_commented_at`: now,
			`citations`:         &p.Citations,
		})

		s.postFullCaches.Delete(int64(ref))
	}

	for _, ref := range refs.Posts.From {
		p, err := s.getPostCached(ctx, int(ref))
		if err != nil {
			if taorm.IsNotFoundError(err) {
				continue
			}
			panic(err)
		}
		if p.Citations.Posts == nil {
			panic(`不应该为 nil`)
		}
		to := &p.Citations.Posts.To
		*to = slices.DeleteFunc(*to, func(n int32) bool { return n == self })
		log.Printf(`删除引用：%d ← %d`, self, ref)

		s.tdb.Model(p).MustUpdateMap(taorm.M{
			`last_commented_at`: now,
			`citations`:         &p.Citations,
		})

		s.deletePostContentCacheFor(int64(ref))
	}
}

// SetPostStatus sets post status.
// 会总是更新 LastCommentedAt 时间。
// TODO 改成内部调用 UpdatePost，并检查 status 是否合法。
func (s *Service) SetPostStatus(ctx context.Context, in *proto.SetPostStatusRequest) (*proto.SetPostStatusResponse, error) {
	user.MustBeAdmin(ctx)

	s.MustTxCall(func(s *Service) error {
		var post models.Post
		s.tdb.Select("id").Where("id=?", in.Id).MustFind(&post)

		if !slices.Contains([]string{
			models.PostStatusPublic,
			models.PostStatusPartial,
			models.PostStatusPrivate,
			models.PostStatusDraft,
		}, in.Status) {
			return errors.New(`无效状态`)
		}

		m := map[string]any{
			"status": in.Status,
		}

		now := time.Now().Unix()

		if in.Touch {
			m[`date`] = now
			m[`modified`] = now
		}

		m[`last_commented_at`] = now

		s.tdb.Model(&post).MustUpdateMap(m)

		s.deletePostContentCacheFor(post.ID)

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

// 最后更新的文章在最后。
func (s *Service) getUserTopPosts(id int) []int {
	if id <= 0 {
		return nil
	}
	var posts []int
	j := utils.Must1(s.options.GetStringDefault(fmt.Sprintf(`user_top_posts:%d`, id), `[]`))
	json.Unmarshal([]byte(j), &posts)
	return posts
}

// TODO 没限制最多数量。
func (s *Service) updateUserTopPosts(id int, postID int, top bool) {
	old := s.getUserTopPosts(id)
	updated := false
	if top {
		// 更新文章的时候如果已经置顶过了，这个列表的顺序不会变。
		if !slices.Contains(old, postID) {
			old = append(old, postID)
			updated = true
		}
	} else {
		if slices.Contains(old, postID) {
			old = slices.DeleteFunc(old, func(p int) bool { return p == postID })
			updated = true
		}
	}
	if updated {
		s.options.SetString(fmt.Sprintf(`user_top_posts:%d`, id), string(utils.Must1(json.Marshal(old))))
	}
}

// 无需鉴权，看不到的文章就是看不到。
func (s *Service) ReorderTopPosts(ctx context.Context, in *proto.ReorderTopPostsRequest) (_ *empty.Empty, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)

	old := s.getUserTopPosts(int(ac.User.ID))
	if len(old) != len(in.Ids) {
		panic(status.Errorf(codes.InvalidArgument, "无效的置顶文章列表。"))
	}

	slices.Reverse(in.Ids)

	s.options.SetString(fmt.Sprintf(`user_top_posts:%d`, ac.User.ID), string(utils.Must1(json.Marshal(in.Ids))))

	return &empty.Empty{}, nil
}

// TODO 不需要公开 api
func (s *Service) GetTopPosts(ctx context.Context, in *proto.GetTopPostsRequest) (*proto.GetTopPostsResponse, error) {
	ac := user.Context(ctx)
	if ac.User.IsGuest() {
		return &proto.GetTopPostsResponse{}, nil
	}
	// 依次调用 GetPost 来获取：
	// 0. 由于是独立维护的，可能有脏数据。
	// 1. 可以判断权限（如果发生变更）
	//    1. 包含文章转移后（没清理干净？）
	//    2. 分享权限发生变更
	// 2. 比 List 的时候 IN (ids) 更快
	posts := []*proto.Post{}
	ids := s.getUserTopPosts(int(ac.User.ID))
	for _, id := range ids {
		p, err := s.GetPost(ctx, &proto.GetPostRequest{
			Id:             int32(id),
			GetPostOptions: in.GetPostOptions,
		})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				continue
			}
			return nil, err
		}
		posts = append(posts, p)
	}

	// ID 是反向保存的，所以要反转。
	slices.Reverse(posts)

	return &proto.GetTopPostsResponse{Posts: posts}, nil
}

// 由于“相关文章”目前只在 GetPost 时返回，所以不在这里设置。
func (s *Service) setPostExtraFields(ctx context.Context, opts *proto.GetPostOptions) func(c *proto.Post) error {
	ac := user.Context(ctx)

	if opts == nil {
		opts = &proto.GetPostOptions{}
	}
	if opts.ContentOptions == nil {
		opts.ContentOptions = &proto.PostContentOptions{}
	}

	topPosts := s.getUserTopPosts(int(ac.User.ID))

	return func(p *proto.Post) error {
		// 私有地址仅对作者可见。
		if ac.User.ID != int64(p.UserId) {
			if p.Metas != nil && p.Metas.Geo != nil && p.Metas.Geo.Private {
				p.Metas.Geo = nil
			}
		}

		if opts.ContentOptions.WithContent {
			content, err := s.getPostContentCached(ctx, p, opts.ContentOptions)
			if err != nil {
				return err
			}
			p.Content = content
		}

		// 碎碎念可能没有标题，自动生成
		//
		// 关于为什么没有在创建/更新的时候生成标题？
		// - 生成算法在变化，而如果保存起来的话，算法变化后不能及时更新，除非全盘重新扫描
		if p.Type == `tweet` {
			switch p.Title {
			case ``, `Untitled`, models.Untitled:
				content, err := s.getPostContentCached(ctx, p, co.For(co.GenerateTweetTitle))
				if err != nil {
					return err
				}
				p.Title = truncateTitle(content, 36)
				p.TitleIsAutoGenerated = true
			}
		}

		if tags, err := s.getPostTagsCached(ctx, p.Id); err != nil {
			return err
		} else {
			p.Tags = tags
		}

		if opts.WithLink != proto.LinkKind_LinkKindNone {
			s.setPostLink(p, opts.WithLink)
		}

		if opts.WithRelates {
			relates, err := s.getRelatedPostsCached(ctx, int64(p.Id))
			if err != nil {
				return err
			}
			for _, r := range relates {
				p.Relates = append(p.Relates, &proto.Post{
					Id:    r.Id,
					Title: r.Title,
				})
			}
			if opts.WithLink != proto.LinkKind_LinkKindNone {
				for _, p := range p.Relates {
					s.setPostLink(p, opts.WithLink)
				}
			}
		}

		if opts.WithComments {
			list, err := s.getPostComments(ctx, p.Id)
			if err != nil {
				return err
			}
			p.CommentList = list
		}

		if opts.WithUserPerms {
			ac := user.MustNotBeGuest(ctx)
			userPerms := utils.Must1(s.GetPostACL(
				user.SystemForLocal(ctx),
				&proto.GetPostACLRequest{PostId: int64(p.Id)}),
			).Users
			canRead := func(userID int32) bool {
				if p, ok := userPerms[userID]; ok {
					return slices.Contains(p.Perms, proto.Perm_PermRead)
				}
				return false
			}
			allUsers := utils.Must1(s.userManager.ListUsers(
				user.SystemForLocal(ctx),
				&proto.ListUsersRequest{}),
			).Users
			allUsers = slices.DeleteFunc(allUsers, func(u *proto.User) bool {
				return u.Id == ac.User.ID
			})
			p.UserPerms = utils.Map(allUsers, func(u *proto.User) *proto.Post_UserPerms {
				return &proto.Post_UserPerms{
					UserId:   int32(u.Id),
					UserName: u.Nickname,
					CanRead:  canRead(int32(u.Id)),
				}
			})
		}

		// TODO 再根据主题设置决定要不要全局开。
		if (opts.WithToc == 1 && p.Metas.Toc) || opts.WithToc == 2 {
			p.Toc = s.getPostTocCached(int(p.Id), p.Source)
		}

		p.Top = slices.Contains(topPosts, int(p.Id))

		return nil
	}
}

// TODO：可能把 [图片] 这种截断
func truncateTitle(title string, length int) string {
	runes := []rune(title)

	for len(runes) > 0 && runes[0] == '\n' {
		runes = runes[1:]
	}

	// 不包含回车、省略号
	if p := slices.IndexFunc(runes, func(r rune) bool {
		switch r {
		case '。', '…', '！':
			return true
		default:
			return false
		}
	}); p > 0 {
		runes = runes[:p]
	}

	// 不超过指定的字符串长度
	maxLength := utils.IIF(length > len(runes), len(runes), length)

	// 不包含句号
	if p := slices.Index(runes, '。'); p > 0 && p < maxLength {
		maxLength = p
	}

	suffix := utils.IIF(len(runes) > maxLength, "...", "")
	return strings.TrimSpace(string(runes[:maxLength]) + suffix)
}

// 请保持文章和评论的代码同步。
func (s *Service) CheckPostTaskListItems(ctx context.Context, in *proto.CheckTaskListItemsRequest) (*proto.CheckTaskListItemsResponse, error) {
	user.MustNotBeGuest(ctx)

	p, err := s.GetPost(ctx,
		&proto.GetPostRequest{
			Id: in.Id,
			GetPostOptions: &proto.GetPostOptions{
				ContentOptions: co.For(co.CheckPostTaskListItems),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	updated, err := s.applyTaskChecks(p.Modified, p.SourceType, p.Source, in)
	if err != nil {
		return nil, err
	}

	p.Source = string(updated)

	updateRequest := proto.UpdatePostRequest{
		Post: p,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				`source_type`,
				`source`,
			},
		},
	}
	updatedPost, err := s.UpdatePost(ctx, &updateRequest)
	if err != nil {
		return nil, err
	}

	return &proto.CheckTaskListItemsResponse{
		ModificationTime: updatedPost.Modified,
	}, nil
}

func (s *Service) applyTaskChecks(modified int32, sourceType, rawSource string, in *proto.CheckTaskListItemsRequest) (string, error) {
	if modified != in.ModificationTime {
		return "", status.Error(codes.Aborted, `内容的修改时间不匹配。`)
	}
	if sourceType != `markdown` {
		return "", status.Error(codes.FailedPrecondition, `内容的类型不支持任务列表。`)
	}

	source := []byte(rawSource)

	apply := func(pos int32, check bool) {
		if pos <= 0 || int(pos) >= len(source)-1 {
			panic(`无效任务。`)
		}
		if (source)[pos-1] != '[' || source[pos+1] != ']' {
			panic(`无效任务。`)
		}
		checked := source[pos] == 'x' || source[pos] == 'X'
		if checked == check {
			panic(`任务状态一致，不能变更。`)
		}
		source[pos] = utils.IIF[byte](check, 'X', ' ')
	}

	if err := (func() (err error) {
		defer utils.CatchAsError(&err)
		for _, item := range in.Checks {
			apply(item, true)
		}
		for _, item := range in.Unchecks {
			apply(item, false)
		}
		return nil
	})(); err != nil {
		return "", err
	}

	return string(source), nil
}

func (s *Service) CreateStylingPage(ctx context.Context, in *proto.CreateStylingPageRequest) (*proto.CreateStylingPageResponse, error) {
	user.MustBeAdmin(ctx)

	source := in.Source
	if source == `` {
		source = string(utils.Must1(fs.ReadFile(styling.Root, `index.md`)))
	}

	id, err := s.options.GetInteger(`styling_page_id`)
	if err != nil {
		if !taorm.IsNotFoundError(err) {
			return nil, err
		}
		var p *proto.Post
		p, err = s.CreatePost(ctx, &proto.Post{
			Title:      `测试页面📄`,
			Slug:       `styling`,
			Type:       `page`,
			Status:     `public`,
			SourceType: `markdown`,
			Source:     source,
		})
		if err == nil {
			s.options.SetInteger(`styling_page_id`, p.Id)
			s.postDataFS.Register(int(p.Id), styling.Root)
		}
	} else {
		var p *proto.Post
		p, err = s.GetPost(ctx, &proto.GetPostRequest{Id: int32(id)})
		if err != nil {
			return nil, err
		}
		_, err = s.UpdatePost(ctx, &proto.UpdatePostRequest{
			Post: &proto.Post{
				Id:         id,
				Title:      `测试页面📄`,
				Modified:   p.Modified,
				Slug:       `styling`,
				SourceType: `markdown`,
				Source:     source,
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					`slug`,
					`source_type`,
					`source`,
					`title`,
				},
			},
		})
	}
	return &proto.CreateStylingPageResponse{}, err
}

func (s *Service) SetPostACL(ctx context.Context, in *proto.SetPostACLRequest) (*proto.SetPostACLResponse, error) {
	// TODO 临时
	user.MustBeAdmin(ctx)

	// 发通知用。
	// 为了取得自动生成的标题，不要使用 getPostCached.
	post := utils.Must1(s.GetPost(
		user.SystemForLocal(context.Background()),
		&proto.GetPostRequest{
			Id: int32(in.PostId),
		},
	))
	owner := utils.Must1(s.userManager.GetUserByID(context.Background(), int(post.UserId)))

	return &proto.SetPostACLResponse{}, s.TxCall(func(s *Service) error {
		// 获取当前的。
		var acl []*models.AccessControlEntry
		s.tdb.Where(`post_id=?`, in.PostId).MustFind(&acl)

		type ACE struct {
			UserID int32
			Perm   proto.Perm
		}

		var old, new []ACE

		for _, ace := range acl {
			var perm proto.Perm
			switch ace.Permission {
			default:
				panic(`错误的权限。`)
			case models.PermRead:
				perm = proto.Perm_PermRead
			}
			old = append(old, ACE{UserID: int32(ace.UserID), Perm: perm})
		}

		for uid, up := range in.Users {
			for _, p := range up.Perms {
				if p == proto.Perm_PermUnknown {
					panic(`错误的权限。`)
				}
				new = append(new, ACE{UserID: uid, Perm: p})
			}
		}

		ps := func(p proto.Perm) string {
			switch p {
			case proto.Perm_PermRead:
				return models.PermRead
			default:
				panic(`无效权限。`)
			}
		}

		for _, a := range old {
			if !slices.Contains(new, a) {
				s.tdb.From(models.AccessControlEntry{}).Where(`user_id=? AND permission=?`, a.UserID, ps(a.Perm)).MustDelete()
			}
		}
		for _, b := range new {
			if !slices.Contains(old, b) {
				ace := models.AccessControlEntry{
					CreatedAt:  time.Now().Unix(),
					PostID:     in.PostId,
					UserID:     int64(b.UserID),
					Permission: ps(b.Perm),
				}
				s.tdb.Model(&ace).MustCreate()

				if in.SendUserNotify {
					// 发送通知。
					// TODO 异步任务
					to := utils.Must1(s.userManager.GetUserByID(db.WithContext(ctx, s.tdb), int(b.UserID)))
					go func() {
						// 仅在分享权限下通知。
						// NOTE：如果更改了受众，但是权限又不是partial，下次更改为 partial 时会丢失分享通知。
						// 因为权限位和受众是独立存储的。
						if post.Status != models.PostStatusPartial {
							return
						}

						// 假装延时一下，以把“新文章发表”通知提前。
						time.Sleep(time.Second * 5)

						// TODO s 内部有 db 事务
						// 异步的时候 goroutine 会拷贝 s 导致事务已提交
						// 所以部分代码放在了 go 之外。
						u := utils.Must1(url.Parse(s.getHome())).JoinPath(s.plainLink(post.Id)).String()
						s.notifier.SendInstant(
							user.SystemForLocal(context.Background()),
							&proto.SendInstantRequest{
								Title: `分享了新文章`,
								Body:  fmt.Sprintf("文章：%s\n来源：%s\n链接：%s", post.Title, owner.Nickname, u),
								// TODO: 没判断为空的情况。如果为空，则分享给了站长。
								BarkToken: to.BarkToken,
							},
						)
					}()
				}
			}
		}

		// 确保文章修改时间更新，方便同步任何检测到文章权限变化。
		s.tdb.MustExec(`UPDATE posts SET modified=? WHERE id=?`, time.Now().Unix(), in.PostId)

		return nil
	})
}

func (s *Service) GetPostACL(ctx context.Context, in *proto.GetPostACLRequest) (*proto.GetPostACLResponse, error) {
	// TODO 临时
	user.MustBeAdmin(ctx)

	var acl []models.AccessControlEntry
	s.tdb.Where(`post_id=?`, in.PostId).MustFind(&acl)

	users := map[int32]*proto.UserPerm{}
	for _, ace := range acl {
		if _, ok := users[int32(ace.UserID)]; !ok {
			users[int32(ace.UserID)] = &proto.UserPerm{}
		}
		var perm proto.Perm
		switch ace.Permission {
		default:
			return nil, errors.New(`错误的权限。`)
		case models.PermRead:
			perm = proto.Perm_PermRead
		}
		users[int32(ace.UserID)].Perms = append(users[int32(ace.UserID)].Perms, perm)
	}

	return &proto.GetPostACLResponse{Users: users}, nil
}

// 快速判断非文章本人用户是否有权限访问被分享的文章。
// NOTE：用于替代 GetPost (withUserPerms)，以提高性能。
// NOTE：系统管理员始终有权限访问。
// NOTE：判断的是**非本人**，本人访问文章不能调用此函数判断。
// NOTE：需保证前提：文章是分享状态。
// TODO：加缓存
func (s *Service) canNonAuthorUserReadPost(ctx context.Context, uid int64, pid int) bool {
	if uid == int64(user.SystemID) {
		return true
	}

	var acl []models.AccessControlEntry
	tdb := db.FromContextDefault(ctx, s.tdb)
	tdb.Where(`post_id=?`, pid).MustFind(&acl)

	return slices.ContainsFunc(acl, func(ace models.AccessControlEntry) bool {
		return ace.UserID == uid && ace.Permission == models.PermRead
	})
}

// 将文章转移到用户名下。
//
// NOTE: 仅管理员可操作。
//
// 前置条件：
func (s *Service) SetPostUserID(ctx context.Context, in *proto.SetPostUserIDRequest) (_ *proto.SetPostUserIDResponse, outErr error) {
	user.MustBeAdmin(ctx)

	s.MustTxCallNoError(func(s *Service) {
		p := utils.Must1(s.getPostCached(ctx, int(in.PostId)))
		if p.UserID == in.UserId {
			panic(`当前用户已是文章作者，无需转移。`)
		}

		// Update() 接口会理想这个字段吗？不确定，先备份一下。
		oldUserID := p.UserID

		// 确保用户存在。
		// 应该 LOCK FOR UPDATE
		utils.Must1(s.userManager.GetUserByID(db.WithContext(ctx, s.tdb), int(in.UserId)))

		// 分类是原作者自己的，不能转移。
		// 自动设置成“未分类”。
		newCategory := 0

		// 没有使用 UpdatePost 函数，有事务冲突。
		s.tdb.Model(p).MustUpdateMap(taorm.M{
			`modified`: time.Now().Unix(),
			`user_id`:  in.UserId,
			`category`: newCategory,
		})

		// 修改权限列表。
		// 1. 把新作者从已有权限中删去（如果有的话）
		// 2. 无需把原作者添加到权限列表。
		acl := utils.Must1(s.GetPostACL(ctx, &proto.GetPostACLRequest{PostId: p.ID}))

		_, sharedToNewUser := acl.Users[in.UserId]
		onlyShare := len(acl.Users) == 1

		delete(acl.Users, in.UserId)
		utils.Must1(s.SetPostACL(ctx, &proto.SetPostACLRequest{
			PostId: p.ID,
			Users:  acl.Users,
		}))

		shouldRemoveTop := false

		// 如果是部分可见且仅分享过给新作者，则设置为私有。
		if p.Status == models.PostStatusPartial && sharedToNewUser && onlyShare {
			utils.Must1(s.SetPostStatus(
				db.WithContext(ctx, s.tdb),
				&proto.SetPostStatusRequest{
					Id:     p.ID,
					Status: models.PostStatusPrivate,
					Touch:  false,
				},
			))
			shouldRemoveTop = true
		} else if p.Status == models.PostStatusPrivate {
			shouldRemoveTop = true
		}

		if shouldRemoveTop {
			s.updateUserTopPosts(int(oldUserID), int(p.ID), false)
		}
	})

	s.deletePostContentCacheFor(in.PostId)
	s.updatePostMetadataTime(in.PostId, time.Now())

	return &proto.SetPostUserIDResponse{}, nil
}

func (s *Service) isPostPublic(ctx context.Context, pid int) bool {
	p, err := s.getPostCached(ctx, pid)
	if err != nil {
		log.Println(err)
		return false
	}
	return p.Status == models.PostStatusPublic
}
