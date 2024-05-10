package service

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SetCommentExtraFieldsContext struct {
	DoNotRenderCodeAsHtml bool
	PrettifyHtml          bool
}

func (s *Service) setCommentExtraFields(ctx context.Context) func(c *protocols.Comment) {
	ac := auth.Context(ctx)

	return func(c *protocols.Comment) {
		c.IsAdmin = s.isAdminEmail(c.Email)
		c.Avatar = int32(s.avatarCache.ID(c.Email))

		// （同 IP 用户 & 5️⃣分钟内） 可编辑。
		// TODO: IP：并不严格判断，比如网吧、办公室可能具有相同 IP。所以限制了时间范围。
		// NOTE：管理员总是可以编辑，跟此值无关。
		// 只允许编辑 Markdown 评论。
		// TODO：其实也允许/也已经支持编辑早期的 HTML 评论，但是在保存的时候已经被转换成 Markdown。
		c.CanEdit = c.SourceType == `markdown` && (ac.RemoteAddr.String() == c.Ip && in5min(c.Date))

		if !ac.User.IsAdmin() {
			c.Email = ""
			c.Ip = ""
		} else {
			c.GeoLocation = s.geoLocation(c.Ip)
		}

		var renderOptions []renderers.Option2
		if ctx, ok := ctx.Value(SetCommentExtraFieldsContext{}).(*SetCommentExtraFieldsContext); ok {
			if ctx.DoNotRenderCodeAsHtml {
				renderOptions = append(renderOptions, renderers.WithDoNotRenderCodeAsHTML())
			}
			if ctx.PrettifyHtml {
				renderOptions = append(renderOptions, renderers.WithHtmlPrettifier())
			}
		}

		switch c.SourceType {
		case `markdown`:
			content, err := s.convertCommentMarkdown(true, c.Source, c.PostId, renderOptions...)
			if err != nil {
				log.Println("转换评论时出错：", err)
				// 也不能干啥……
				// fallthrough
			}
			c.Content = content
		case `plain`:
			c.Content = c.Source
		default:
			c.Content = ""
		}
	}
}

// Deprecated.
func (s *Service) comments() *taorm.Stmt {
	return s.tdb.Model(models.Comment{})
}

// GetComment2 ...
func (s *Service) getComment2(name int64) *models.Comment {
	var comment models.Comment
	s.comments().Where("id=?", name).MustFind(&comment)
	return &comment
}

// 像二狗说的那样，服务启动时缓存所有的头像哈希值，
// 否则缓存的页面图片在服务重启后刷新时会加载失败。
// https://qwq.me/p/249/1#comment-506
// NOTE: ORM 不支持 distinct，所以没写。
func (s *Service) cacheAllCommenterData() {
	var comments models.Comments
	s.tdb.Select(`email,ip`).OrderBy(`date desc`).MustFind(&comments)
	for _, c := range comments {
		_ = s.avatarCache.ID(c.Email)
	}
	if !strings.EqualFold(version.GitCommit, `head`) {
		go func() {
			for _, c := range comments {
				s.cmtgeo.Queue(c.IP, nil)
			}
		}()
	}
}

// GetComment ...
// TODO perm check
// TODO remove email & user
func (s *Service) GetComment(ctx context.Context, req *protocols.GetCommentRequest) (*protocols.Comment, error) {
	return s.getComment2(req.Id).ToProtocols(s.setCommentExtraFields(ctx)), nil
}

func in5min(t int32) bool {
	return time.Since(time.Unix(int64(t), 0)) < time.Minute*5
}

// 更新评论。
//
// NOTE：只支持更新评论内容。
// NOTE：带上时间戳，防止异地多次更新的冲突（太严格了吧！）
func (s *Service) UpdateComment(ctx context.Context, req *protocols.UpdateCommentRequest) (*protocols.Comment, error) {
	ac := auth.Context(ctx)
	cmtOld := s.getComment2(req.Comment.Id)
	if !ac.User.IsAdmin() {
		if ac.RemoteAddr.String() != cmtOld.IP || !in5min(cmtOld.Date) {
			return nil, status.Error(codes.PermissionDenied, `超时或无权限编辑评论`)
		}
	}

	var comment models.Comment

	if req.Comment != nil && req.UpdateMask != nil && req.UpdateMask.Paths != nil {
		data := map[string]interface{}{}
		var hasSourceType, hasSource bool
		var hasModified bool
		for _, mask := range req.UpdateMask.Paths {
			switch mask {
			default:
				panic(`unknown update_mask field`)
			case `source_type`:
				hasSourceType = true
				data[`source_type`] = req.Comment.SourceType
			case `source`:
				hasSource = true
				data[`source`] = req.Comment.Source
			case `modified`:
				hasModified = true
				data[`modified`] = time.Now().Unix()
			}
		}
		if !hasModified {
			return nil, status.Error(codes.Aborted, `更新评论需要带上评论本身的修改时间。`)
		} else if cmtOld.Modified != req.Comment.Modified {
			return nil, status.Error(codes.Aborted, `当前评论内容已在其它地方被修改过，请刷新后重新提交。`)
		}
		if hasSourceType != hasSource {
			panic(`source_type and source must be both specified`)
		}
		if hasSourceType {
			if strings.TrimSpace(req.Comment.Source) == "" {
				return nil, status.Error(codes.InvalidArgument, `评论内容不能为空。`)
			}
			if req.Comment.SourceType != `markdown` {
				return nil, status.Error(codes.InvalidArgument, `不再允许发表非 Markdown 评论。`)
			}
			if _, err := s.convertCommentMarkdown(ac.User.IsAdmin(), req.Comment.Source, cmtOld.PostID, renderers.WithoutRendering()); err != nil {
				return nil, err
			}
		}
		s.MustTxCall(func(txs *Service) error {
			txs.tdb.Model(models.Comment{}).Where(`id=?`, req.Comment.Id).MustUpdateMap(data)
			txs.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
			txs.updatePostCommentCount(comment.PostID, time.Now())
			return nil
		})
	} else {
		s.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
	}

	return comment.ToProtocols(s.setCommentExtraFields(ctx)), nil
}

// DeleteComment ...
func (s *Service) DeleteComment(ctx context.Context, in *protocols.DeleteCommentRequest) (*protocols.DeleteCommentResponse, error) {
	s.MustBeAdmin(ctx)
	cmt := s.getComment2(int64(in.Id))
	s.comments().Where("id=? OR root=?", in.Id, in.Id).Delete()
	s.updatePostCommentCount(cmt.PostID, time.Now())
	s.updateCommentsCount()
	return &protocols.DeleteCommentResponse{}, nil
}

// ListComments ...
func (s *Service) ListComments(ctx context.Context, in *protocols.ListCommentsRequest) (*protocols.ListCommentsResponse, error) {
	ac := auth.Context(ctx)

	if in.Limit <= 0 || in.Limit > 100 {
		panic(status.Errorf(codes.InvalidArgument, `limit out of range`))
	}

	if in.Mode == protocols.ListCommentsRequest_Unspecified {
		in.Mode = protocols.ListCommentsRequest_Tree
	}

	var parents models.Comments
	{
		// TODO ensure that fields must include root etc to be used later.
		// TODO verify fields that are sanitized.
		stmt := s.tdb.Select(strings.Join(in.Fields, ","))
		stmt.WhereIf(in.Mode == protocols.ListCommentsRequest_Tree, "root = 0")
		stmt.WhereIf(in.PostId > 0, "post_id=?", in.PostId)
		// limit & offset apply to parent comments only
		stmt.Limit(in.Limit).Offset(in.Offset).OrderBy(in.OrderBy)
		if ac.User.IsGuest() {
			stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
		}
		if len(in.Types) > 0 {
			stmt.InnerJoin(`posts`, `comments.post_id = posts.id AND posts.type in ?`, in.Types)
		}
		stmt.MustFind(&parents)
	}

	var children models.Comments

	// 其实是可以合并这两段高度相似的代码的，不过，因为 limit/offset 只限制顶级评论不限制子评论的原因，SQL 语句不好写。
	if in.Mode == protocols.ListCommentsRequest_Tree && len(parents) > 0 {
		parentIDs := make([]int64, 0, len(parents))
		for _, parent := range parents {
			parentIDs = append(parentIDs, parent.ID)
		}

		stmt := s.tdb.Select(strings.Join(in.Fields, ","))
		stmt.Where("root IN (?)", parentIDs)
		if ac.User.IsGuest() {
			stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
		}
		if len(in.Types) > 0 {
			stmt.InnerJoin(`posts`, `comments.post_id = posts.id AND posts.type in ?`, in.Types)
		}
		stmt.MustFind(&children)
	}

	comments := make(models.Comments, 0, len(parents)+len(children))
	comments = append(comments, parents...)
	comments = append(comments, children...)

	ctx = context.WithValue(ctx, SetCommentExtraFieldsContext{}, &SetCommentExtraFieldsContext{
		DoNotRenderCodeAsHtml: in.DoNotRenderCodeAsHtml,
		PrettifyHtml:          in.PrettifyHtml,
	})

	protoComments := comments.ToProtocols(s.setCommentExtraFields(ctx))

	return &protocols.ListCommentsResponse{Comments: protoComments}, nil
}

func (s *Service) GetPostComments(ctx context.Context, req *protocols.GetPostCommentsRequest) (*protocols.GetPostCommentsResponse, error) {
	ac := auth.Context(ctx)
	if !(ac.User.IsAdmin() || s.isPostPublic(ctx, req.Id)) {
		return nil, status.Error(codes.PermissionDenied, `你无权查看此文章的评论。`)
	}
	var comments models.Comments
	stmt := s.tdb.Select(`*`)
	stmt.Where("post_id=?", req.Id)
	stmt.MustFind(&comments)
	return &protocols.GetPostCommentsResponse{
		Comments: comments.ToProtocols(s.setCommentExtraFields(ctx)),
	}, nil
}

func (s *Service) GetAllCommentsCount() int64 {
	var count int64
	s.tdb.Model(models.Comment{}).Select("count(1) as count").Find(&count)
	return count
}

func (s *Service) geoLocation(ip string) string {
	go func() {
		if err := s.cmtgeo.Queue(ip, nil); err != nil {
			log.Println(`GeoLocation.Queue:`, ip, err)
		}
	}()
	return s.cmtgeo.Get(ip)
}

const (
	maxNicknameLen = 32
	maxEmailLen    = 64
	maxUrlLen      = 256
	maxContentLen  = 4096
)

// 创建一条评论。
//
// 前期验证项：
//
// Author 不为空、长度不超限
// Email 不为空、长度不超限
// URL 可为空，若不为空，需要为正确的 URL，可不加 http 前缀，自动加
// SourceType 只能为 markdown
// Source 不为空且不超限。
//
// 检查昵称是否被允许
// 检查邮箱是否被允许
//
// 检查项：
// post id 存在
// parent 是否存在
// 且和 parent 的 post id 一样
// root 置为 parent / parent的root
// IP 从请求中自动获取，忽略传入。
// Date 服务端的当前时间戳，忽略传入。
// Content 自动由 source 生成。
//
// NOTE: 默认的 modified 修改时间为 0，表示从未被修改过。
func (s *Service) CreateComment(ctx context.Context, in *protocols.Comment) (*protocols.Comment, error) {
	ac := auth.Context(ctx)

	// 尽早查询地理信息
	go func() {
		if err := s.cmtgeo.Queue(ac.RemoteAddr.String(), nil); err != nil {
			log.Println(err)
		}
	}()

	now := time.Now()

	c := models.Comment{
		PostID:     in.PostId,
		Parent:     in.Parent,
		Author:     strings.TrimSpace(in.Author),
		Email:      strings.TrimSpace(in.Email),
		URL:        strings.TrimSpace(in.Url),
		IP:         ac.RemoteAddr.String(),
		Date:       0,
		SourceType: in.SourceType,
		Source:     in.Source,
	}

	if in.Modified > 0 {
		c.Modified = in.Modified
	}
	if in.Date > 0 {
		c.Date = in.Date
		if in.Modified == 0 {
			c.Modified = c.Date
		}
	} else {
		c.Date = int32(now.Unix())
		c.Modified = int32(now.Unix())
	}

	if c.Author == "" {
		return nil, status.Error(codes.InvalidArgument, `昵称不能为空`)
	}
	if utf8.RuneCountInString(c.Author) >= maxNicknameLen {
		return nil, status.Errorf(codes.InvalidArgument, `昵称太长（最长 %d 个字符）`, maxNicknameLen)
	}

	if !utils.IsEmail(c.Email) {
		return nil, status.Errorf(codes.InvalidArgument, `邮箱格式不正确`)
	}
	if utf8.RuneCountInString(c.Email) >= maxEmailLen {
		return nil, status.Errorf(codes.InvalidArgument, `邮箱太长（最长 %d 个字符）`, maxEmailLen)
	}

	if c.URL != "" {
		if !strings.Contains(c.URL, "://") {
			c.URL = `http://` + c.URL
		}
		if !utils.IsURL(c.URL, false) {
			return nil, status.Errorf(codes.InvalidArgument, `网址格式不正确`)
		}
	}
	if utf8.RuneCountInString(c.URL) >= maxUrlLen {
		return nil, status.Errorf(codes.InvalidArgument, `网址太长（最长 %d 个字符）`, maxUrlLen)
	}

	if c.SourceType != `markdown` && !ac.User.IsAdmin() {
		return nil, status.Error(codes.InvalidArgument, `不再允许发表非 Markdown 评论。`)
	}
	if strings.TrimSpace(c.Source) == "" {
		return nil, status.Error(codes.InvalidArgument, `评论内容不能为空`)
	}
	if utf8.RuneCountInString(c.Source) >= maxContentLen {
		return nil, status.Error(codes.InvalidArgument, `评论内容太长`)
	}

	if ac.User.IsGuest() {
		notAllowedEmails := s.cfg.Comment.NotAllowedEmails
		if adminEmails := s.cfg.Comment.Emails; len(adminEmails) > 0 {
			notAllowedEmails = append(notAllowedEmails, adminEmails...)
		}
		for _, email := range notAllowedEmails {
			if email != "" && c.Email != "" && strings.EqualFold(email, c.Email) {
				return nil, status.Error(codes.InvalidArgument, `不能使用此邮箱地址`)
			}
		}
		notAllowedAuthors := s.cfg.Comment.NotAllowedAuthors
		if adminName := s.cfg.Comment.Author; adminName != "" {
			notAllowedAuthors = append(notAllowedAuthors, adminName)
		}
		for _, author := range notAllowedAuthors {
			if author != "" && c.Author != "" && strings.EqualFold(author, string(c.Author)) {
				return nil, status.Error(codes.InvalidArgument, `不能使用此昵称`)
			}
		}
		if c.Author != "" && strings.Contains(in.Author, "作者") {
			return nil, status.Error(codes.InvalidArgument, "昵称中不应包含“作者”两字")
		}
	}

	if c.SourceType == `markdown` {
		if _, err := s.convertCommentMarkdown(ac.User.IsAdmin(), c.Source, c.PostID, renderers.WithoutRendering()); err != nil {
			return nil, err
		}
	}

	s.MustTxCall(func(txs *Service) error {
		if c.Parent > 0 {
			pc := txs.getComment2(c.Parent)
			if pc.Root != 0 {
				c.Root = pc.Root
			} else {
				c.Root = pc.ID
			}
			if c.PostID != pc.PostID {
				panic(status.Error(codes.InvalidArgument, `不是同一篇文章的父评论。`))
			}
		} else {
			c.Root = 0
		}
		txs.tdb.Model(&c).MustCreate()
		txs.updatePostCommentCount(c.PostID, time.Unix(int64(c.Date), 0))
		txs.updateCommentsCount()
		return nil
	})

	s.doCommentNotification(&c)

	return c.ToProtocols(s.setCommentExtraFields(ctx)), nil
}

func (s *Service) updateCommentsCount() {
	count := s.GetAllCommentsCount()
	s.SetOption("comment_count", count)
}

// 发表/更新评论时：普通用户不能发表 HTML 评论，管理员可以。
// 一旦发表/更新成功：始终认为评论是合法的。
//
// 换言之，发表/更新调用此接口，把评论转换成 html 时用 convert 接口。
// 前者用请求身份，后者不限身份。
// TODO 像 getPostContent 一样走缓存。
func (s *Service) convertCommentMarkdown(secure bool, source string, postID int64, options ...renderers.Option2) (string, error) {
	opts := []renderers.Option2{
		renderers.WithPathResolver(s.PathResolver(postID)),
		renderers.WithOpenLinksInNewTab(renderers.OpenLinksInNewTabKindExternal),
	}
	if !secure {
		opts = append(opts,
			renderers.WithDisableHeadings(true),
			renderers.WithDisableHTML(true),
		)
	}

	// 最后才追加，允许外部覆盖内部默认。
	opts = append(opts, options...)

	md := renderers.NewMarkdown(opts...)

	_, content, err := md.Render(source)
	if err != nil {
		return ``, fmt.Errorf(`转换 Markdown 时出错：%w`, err)
	}

	return content, nil
}

// SetCommentPostID 把某条顶级评论及其子评论转移到另一篇文章下
func (s *Service) SetCommentPostID(ctx context.Context, in *protocols.SetCommentPostIDRequest) (*protocols.SetCommentPostIDResponse, error) {
	s.MustBeAdmin(ctx)

	s.MustTxCall(func(txs *Service) error {
		cmt := txs.getComment2(in.Id)
		if cmt.Root != 0 {
			panic(`不能转移子评论`)
		}
		post := txs.GetPostByID(ctx, in.PostId)
		if cmt.PostID == post.Id {
			panic(`不能转移到相同的文章`)
		}
		txs.tdb.From(cmt).
			Where(`post_id=?`, cmt.PostID).
			Where(`id=? OR root=?`, cmt.ID, cmt.ID).
			MustUpdateMap(map[string]interface{}{
				`post_id`: post.Id,
			})
		txs.updatePostCommentCount(cmt.PostID, time.Now())
		txs.updatePostCommentCount(post.Id, time.Now())
		log.Printf("Transferred comments %d to post %d", cmt.ID, in.PostId)
		return nil
	})

	return &protocols.SetCommentPostIDResponse{}, nil
}

func (s *Service) PreviewComment(ctx context.Context, in *protocols.PreviewCommentRequest) (*protocols.PreviewCommentResponse, error) {
	ac := auth.Context(ctx)

	options := []renderers.Option2{}
	if in.OpenLinksInNewTab {
		options = append(options, renderers.WithOpenLinksInNewTab(renderers.OpenLinksInNewTabKindAll))
	}

	// TODO 安全检查：PostID 应该和 Referer 一致。
	content, err := s.convertCommentMarkdown(ac.User.IsAdmin(), in.Markdown, int64(in.PostId), options...)
	return &protocols.PreviewCommentResponse{Html: content}, err
}

// 判断评论者的邮箱是否为管理员。
// 不区分大小写。
func (s *Service) isAdminEmail(email string) bool {
	return slices.ContainsFunc(s.cfg.Comment.Emails, func(s string) bool {
		return strings.EqualFold(email, s)
	})
}

// TODO 改成后台任务异步通知（因为可能失败，失败后应该重试）
func (s *Service) doCommentNotification(c *models.Comment) {
	if !s.cfg.Comment.Notify {
		log.Printf(`comment notification is disabled. comment_id: %v, post_id: %v`, c.ID, c.PostID)
		return
	}

	postTitle := s.GetPostTitle(c.PostID)
	// TODO 修改链接。
	postLink := fmt.Sprintf("%s/%d/#comment-%d", s.HomeURL(), c.PostID, c.ID)
	adminEmails := s.cfg.Comment.Emails
	if len(adminEmails) == 0 {
		return
	}

	if !s.isAdminEmail(c.Email) {
		data := &comment_notify.AdminData{
			Title:    postTitle,
			Link:     postLink,
			Date:     time.Unix(int64(c.Date), 0).Local().Format(time.RFC3339),
			Author:   c.Author,
			Content:  c.Source,
			Email:    c.Email,
			HomePage: c.URL,
		}
		s.cmtntf.NotifyAdmin(data)

		if config := s.cmtntf.Config.Push.Chanify; config.Token != "" {
			comment_notify.Chanify(config.Token, data)
		}
	}

	var parents []models.Comment

	for parentID := c.Parent; parentID > 0; {
		var parent models.Comment
		s.tdb.From(parent).
			Select("id,author,email,parent").
			Where("id=?", parentID).
			MustFind(&parent)
		parents = append(parents, parent)
		parentID = parent.Parent
	}

	// not a reply to some comment
	if len(parents) == 0 {
		return
	}

	var distinctNames []string
	var distinctEmails []string

	distinct := map[string]bool{}
	for _, parent := range parents {
		if s.isAdminEmail(parent.Email) || strings.EqualFold(parent.Email, c.Email) {
			continue
		}
		email := strings.ToLower(parent.Email)
		if _, ok := distinct[email]; !ok {
			distinct[email] = true
			distinctNames = append(distinctNames, parent.Author)
			distinctEmails = append(distinctEmails, parent.Email)
		}
	}

	if len(distinctNames) == 0 {
		return
	}

	guestData := comment_notify.GuestData{
		Title:   postTitle,
		Link:    postLink,
		Date:    time.Unix(int64(c.Date), 0).Local().Format(time.RFC3339),
		Author:  c.Author,
		Content: c.Source,
	}

	s.cmtntf.NotifyGuests(&guestData, distinctNames, distinctEmails)
}

func (s *Service) deletePostComments(_ context.Context, postID int64) {
	s.tdb.From(models.Comment{}).Where(`post_id=?`, postID).MustDelete()
}
