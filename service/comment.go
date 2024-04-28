package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
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
	"github.com/movsb/taorm/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

func (s *Service) avatar(email string) int {
	return s.avatarCache.ID(email)
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
	user := s.auth.AuthGRPC(ctx)
	return s.getComment2(req.Id).ToProtocols(s.isAdminEmail, user, s.geoLocation, "", s.avatar), nil
}

// UpdateComment ...
func (s *Service) UpdateComment(ctx context.Context, req *protocols.UpdateCommentRequest) (*protocols.Comment, error) {
	user := s.auth.AuthGRPC(ctx)
	cmtOld := s.getComment2(req.Comment.Id)
	if !user.IsAdmin() {
		userIP := ipFromContext(ctx, true)
		if userIP != cmtOld.IP || !models.In5min(cmtOld.Date) {
			return nil, status.Error(codes.PermissionDenied, `超时或无权限编辑评论`)
		}
	}

	var comment models.Comment

	if req.Comment != nil && req.UpdateMask != nil && req.UpdateMask.Paths != nil {
		data := map[string]interface{}{}
		var hasSourceType, hasSource bool
		for _, mask := range req.UpdateMask.Paths {
			switch mask {
			default:
				panic(`unknown update_mask field`)
			case `source_type`:
				hasSourceType = true
			case `source`:
				hasSource = true
			}
		}
		if hasSourceType != hasSource {
			panic(`source_type and source must be both specified`)
		}
		if hasSourceType {
			data[`source_type`] = req.Comment.SourceType
			data[`source`] = req.Comment.Source

			if content, err := s.convertCommentMarkdown(user, req.Comment.SourceType, req.Comment.Source, cmtOld.PostID); err != nil {
				return nil, err
			} else {
				data[`content`] = content
			}
		}
		s.MustTxCall(func(txs *Service) error {
			txs.tdb.Model(models.Comment{}).Where(`id=?`, req.Comment.Id).MustUpdateMap(data)
			txs.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
			return nil
		})
	} else {
		s.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
	}

	return comment.ToProtocols(s.isAdminEmail, user, s.geoLocation, "", s.avatar), nil
}

// DeleteComment ...
func (s *Service) DeleteComment(ctx context.Context, in *protocols.DeleteCommentRequest) (*protocols.DeleteCommentResponse, error) {
	s.auth.AuthGRPC(ctx).MustBeAdmin()
	cmt := s.getComment2(int64(in.Id))
	s.comments().Where("id=? OR root=?", in.Id, in.Id).Delete()
	s.UpdatePostCommentCount(cmt.PostID)
	return &protocols.DeleteCommentResponse{}, nil
}

// ListComments ...
func (s *Service) ListComments(ctx context.Context, in *protocols.ListCommentsRequest) (*protocols.ListCommentsResponse, error) {
	user := s.auth.AuthGRPC(ctx)
	userIP := ipFromContext(ctx, false)

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
		if user.IsGuest() {
			stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
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
		if user.IsGuest() {
			stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
		}
		stmt.MustFind(&children)
	}

	comments := make(models.Comments, 0, len(parents)+len(children))
	comments = append(comments, parents...)
	comments = append(comments, children...)

	protoComments := comments.ToProtocols(s.isAdminEmail, user, s.geoLocation, userIP, s.avatar)

	return &protocols.ListCommentsResponse{Comments: protoComments}, nil
}

// 用于前端渲染全部评论的函数。
func (s *Service) ListPostAllComments(req *http.Request, pid int64) []*protocols.Comment {
	user := s.auth.AuthRequest(req)
	header := req.Header.Get(`x-forwarded-for`)
	if header == "" {
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		host = strings.Trim(host, `[]`)
		header = host
	}
	userIP := ipFromContext(metadata.NewIncomingContext(req.Context(), metadata.Pairs(`x-forwarded-for`, header)), true)

	var comments models.Comments
	stmt := s.tdb.Select(`*`)
	stmt.Where("post_id=?", pid)
	stmt.MustFind(&comments)

	return comments.ToProtocols(s.isAdminEmail, user, s.geoLocation, userIP, s.avatar)
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

// TODO this is temp
// TODO not http only
func ipFromContext(ctx context.Context, must bool) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		if must {
			panic(`no md`)
		}
		return ``
	}
	var forward string

	if forwards, ok := md["x-forwarded-for"]; ok && len(forwards) > 0 {
		forward = forwards[0]
	}
	if forward == "" {
		if must {
			panic("invalid request") // TODO HTTP 400
		}
		return ``
	}

	// since IP field has no room for proxies, strip them all.
	// https://en.wikipedia.org/wiki/X-Forwarded-For#Format
	// https://github.com/grpc-ecosystem/grpc-gateway/blob/20f268a412e5b342ebfb1a0eef7c3b7bd6c260ea/runtime/context.go#L103
	if p := strings.IndexByte(forward, ','); p != -1 {
		forward = forward[:p]
	}

	return forward
}

const (
	maxNicknameLen = 32
	maxEmailLen    = 64
	maxUrlLen      = 256
	maxContentLen  = 4096
)

// CreateComment ...
func (s *Service) CreateComment(ctx context.Context, in *protocols.Comment) (*protocols.Comment, error) {
	user := s.auth.User(ctx)

	ip := ipFromContext(ctx, true)

	// 尽早查询地理信息
	go func() {
		if err := s.cmtgeo.Queue(ip, nil); err != nil {
			log.Println(err)
		}
	}()

	comment := models.Comment{
		PostID:     in.PostId,
		Parent:     in.Parent,
		Author:     in.Author,
		Email:      in.Email,
		URL:        in.Url,
		IP:         ip,
		Date:       int32(time.Now().Unix()),
		SourceType: in.SourceType,
		Source:     in.Source,
	}

	if in.Author == "" {
		return nil, status.Error(codes.InvalidArgument, `昵称不能为空`)
	}
	if utf8.RuneCountInString(in.Author) >= maxNicknameLen {
		errors.Join()
		return nil, status.Errorf(codes.InvalidArgument, `昵称太长（最长 %d 个字符）`, maxNicknameLen)
	}

	if !utils.IsEmail(in.Email) {
		return nil, status.Errorf(codes.InvalidArgument, `邮箱格式不正确`)
	}
	if utf8.RuneCountInString(in.Email) >= maxEmailLen {
		return nil, status.Errorf(codes.InvalidArgument, `邮箱太长（最长 %d 个字符）`, maxEmailLen)
	}

	if in.Url != "" && !utils.IsURL(in.Url, true) {
		return nil, status.Errorf(codes.InvalidArgument, `网址格式不正确`)
	}
	if utf8.RuneCountInString(in.Url) >= maxUrlLen {
		return nil, status.Errorf(codes.InvalidArgument, `网址太长（最长 %d 个字符）`, maxUrlLen)
	}

	if in.Source == "" {
		return nil, status.Error(codes.InvalidArgument, `评论内容不能为空`)
	}
	if utf8.RuneCountInString(in.Source) >= maxContentLen {
		return nil, status.Error(codes.InvalidArgument, `评论内容太长`)
	}

	if in.Parent > 0 {
		pc := s.getComment2(in.Parent)
		comment.Root = pc.Root
		if pc.Root == 0 {
			comment.Root = pc.ID
		}
	}

	if content, err := s.convertCommentMarkdown(user, in.SourceType, in.Source, in.PostId); err == nil {
		comment.Content = content
	} else {
		return nil, err
	}

	adminEmails := s.cfg.Comment.Emails

	if user.IsGuest() {
		notAllowedEmails := s.cfg.Comment.NotAllowedEmails
		if len(adminEmails) > 0 {
			notAllowedEmails = append(notAllowedEmails, adminEmails...)
		}
		// TODO use regexp to detect equality.
		for _, email := range notAllowedEmails {
			if email != "" && in.Email != "" && strings.EqualFold(email, in.Email) {
				return nil, status.Error(codes.InvalidArgument, `不能使用此邮箱地址`)
			}
		}
		notAllowedAuthors := s.cfg.Comment.NotAllowedAuthors
		if adminName := s.cfg.Comment.Author; adminName != "" {
			notAllowedAuthors = append(notAllowedAuthors, adminName)
		}
		for _, author := range notAllowedAuthors {
			if author != "" && in.Author != "" && strings.EqualFold(author, string(in.Author)) {
				return nil, status.Error(codes.InvalidArgument, `不能使用此昵称`)
			}
		}
		if in.Author != "" && strings.Contains(in.Author, "作者") {
			return nil, status.Error(codes.InvalidArgument, "昵称中不应包含“作者”两字")
		}
	}

	s.MustTxCall(func(txs *Service) error {
		txs.tdb.Model(&comment).MustCreate()
		txs.updateCommentsCount()
		txs.UpdatePostCommentCount(comment.PostID)
		return nil
	})

	s.doCommentNotification(&comment)

	return comment.ToProtocols(s.isAdminEmail, user, s.geoLocation, ip, s.avatar), nil
}

func (s *Service) updateCommentsCount() {
	count := s.GetAllCommentsCount()
	s.SetOption("comment_count", count)
}

func (s *Service) convertCommentMarkdown(user *auth.User, ty string, source string, postID int64, options ...renderers.Option) (string, error) {
	if ty != "markdown" {
		return "", status.Error(codes.InvalidArgument, "仅支持 Markdown")
	}

	var md renderers.Renderer

	options = append(options,
		renderers.WithPathResolver(s.PathResolver(postID)),
	)

	if user.IsAdmin() {
		md = renderers.NewMarkdown(options...)
	} else {
		options = append(options,
			renderers.WithDisableHeadings(true),
			renderers.WithDisableHTML(true),
		)
		md = renderers.NewMarkdown(options...)
	}

	_, content, err := md.Render(source)
	if err != nil {
		return ``, fmt.Errorf(`转换 Markdown 时出错：%w`, err)
	}

	return content, nil
}

// SetCommentPostID 把某条顶级评论及其子评论转移到另一篇文章下
func (s *Service) SetCommentPostID(ctx context.Context, in *protocols.SetCommentPostIDRequest) (*protocols.SetCommentPostIDResponse, error) {
	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() {
		// TODO 没有必要细说，统一成无权限就行了。
		return nil, status.Error(codes.PermissionDenied, `你没有权限转移评论。`)
	}

	s.MustTxCall(func(txs *Service) error {
		cmt := txs.getComment2(in.Id)
		if cmt.Root != 0 {
			panic(`不能转移子评论`)
		}
		post := txs.GetPostByID(in.PostId)
		if cmt.PostID == post.Id {
			panic(`不能转移到相同的文章`)
		}
		txs.tdb.From(cmt).
			Where(`post_id=?`, cmt.PostID).
			Where(`id=? OR root=?`, cmt.ID, cmt.ID).
			MustUpdateMap(map[string]interface{}{
				`post_id`: post.Id,
			})
		txs.UpdatePostCommentCount(cmt.PostID)
		txs.UpdatePostCommentCount(post.Id)
		log.Printf("Transferred comments %d to post %d", cmt.ID, in.PostId)
		return nil
	})

	return &protocols.SetCommentPostIDResponse{}, nil
}

func (s *Service) PreviewComment(ctx context.Context, in *protocols.PreviewCommentRequest) (*protocols.PreviewCommentResponse, error) {
	user := s.auth.AuthGRPC(ctx)
	options := []renderers.Option{}
	if in.OpenLinksInNewTab {
		options = append(options, renderers.WithOpenLinksInNewTab())
	}
	// TODO 安全检查：PostID 应该和 Referer 一致。
	content, err := s.convertCommentMarkdown(user, `markdown`, in.Markdown, int64(in.PostId), options...)
	return &protocols.PreviewCommentResponse{Html: content}, err
}

func (s *Service) isAdminEmail(email string) bool {
	return slices.ContainsFunc(s.cfg.Comment.Emails, func(s string) bool {
		return strings.EqualFold(email, s)
	})
}

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
		if data.Content == "" {
			data.Content = c.Content
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
	if guestData.Content == "" {
		guestData.Content = c.Content
	}

	s.cmtntf.NotifyGuests(&guestData, distinctNames, distinctEmails)
}

func (s *Service) deletePostComments(_ context.Context, postID int64) {
	s.tdb.From(models.Comment{}).Where(`post_id=?`, postID).MustDelete()
}
