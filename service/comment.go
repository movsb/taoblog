package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taorm/taorm"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// Deprecated.
func (s *Service) comments() *taorm.Stmt {
	return s.tdb.Model(models.Comment{})
}

// GetComment2 ...
func (s *Service) GetComment2(name int64) *models.Comment {
	var comment models.Comment
	s.comments().Where("id=?", name).MustFind(&comment)
	return &comment
}

// GetComment ...
// TODO perm check
// TODO remove email & user
func (s *Service) GetComment(ctx context.Context, req *protocols.GetCommentRequest) (*protocols.Comment, error) {
	user := s.auth.AuthGRPC(ctx)
	return s.GetComment2(req.Id).ToProtocols(s.cfg.Comment.Email, user), nil
}

// UpdateComment ...
func (s *Service) UpdateComment(ctx context.Context, req *protocols.UpdateCommentRequest) (*protocols.Comment, error) {
	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() {
		panic(`not enough permission`)
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
			data[`content`] = s.convertCommentMarkdown(req.Comment)
		}
		s.TxCall(func(txs *Service) error {
			txs.tdb.Model(models.Comment{}).Where(`id=?`, req.Comment.Id).MustUpdateMap(data)
			txs.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
			return nil
		})
	} else {
		s.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
	}

	return comment.ToProtocols(s.cfg.Comment.Email, user), nil
}

// ListComments ...
func (s *Service) ListComments(ctx context.Context, in *protocols.ListCommentsRequest) (*protocols.ListCommentsResponse, error) {
	user := s.auth.AuthGRPC(ctx)
	adminEmail := s.cfg.Comment.Email

	var parentProtocolComments []*protocols.Comment
	{
		var parents models.Comments
		// TODO ensure that fields must include root etc to be used later.
		// TODO verify fields that are sanitized.
		stmt := s.tdb.Select(strings.Join(in.Fields, ","))
		stmt.WhereIf(in.Mode == protocols.ListCommentsMode_ListCommentsModeTree, "root = 0")
		stmt.WhereIf(in.PostId > 0, "post_id=?", in.PostId)
		// limit & offset apply to parent comments only
		stmt.Limit(in.Limit).Offset(in.Offset).OrderBy(in.OrderBy)
		if user.IsGuest() {
			stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
		}
		stmt.MustFind(&parents)
		parentProtocolComments = parents.ToProtocols(adminEmail, user)
	}

	needChildren := in.Mode == protocols.ListCommentsMode_ListCommentsModeTree && len(parentProtocolComments) > 0

	if needChildren {
		var childrenProtocolComments []*protocols.Comment
		{
			parentIDs := make([]int64, 0, len(parentProtocolComments))
			for _, parent := range parentProtocolComments {
				parentIDs = append(parentIDs, parent.Id)
			}
			var children models.Comments
			stmt := s.tdb.Select(strings.Join(in.Fields, ","))
			stmt.Where("root IN (?)", parentIDs)
			if user.IsGuest() {
				stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
			}
			stmt.MustFind(&children)
			childrenProtocolComments = children.ToProtocols(adminEmail, user)
		}
		{
			childrenMap := make(map[int64][]*protocols.Comment, len(parentProtocolComments))
			for _, child := range childrenProtocolComments {
				childrenMap[child.Root] = append(childrenMap[child.Root], child)
			}
			for _, parent := range parentProtocolComments {
				parent.Children = childrenMap[parent.Id]
			}
		}
	}

	return &protocols.ListCommentsResponse{
		Comments: parentProtocolComments,
	}, nil
}

func (s *Service) GetAllCommentsCount() int64 {
	var count int64
	s.tdb.Model(models.Comment{}).Select("count(1) as count").Find(&count)
	return count
}

func (s *Service) CreateComment(ctx context.Context, c *protocols.Comment) *protocols.Comment {
	user := s.auth.AuthContext(ctx)

	comment := models.Comment{
		PostID:     c.PostId,
		Parent:     c.Parent,
		Author:     c.Author,
		Email:      c.Email,
		URL:        c.Url,
		IP:         c.Ip,
		Date:       datetime.Proto2My(*c.Date),
		SourceType: c.SourceType,
		Source:     c.Source,
	}

	if c.Author == "" {
		panic(exception.NewValidationError("昵称不能为空"))
	}

	if utf8.RuneCountInString(c.Author) >= 32 {
		panic(exception.NewValidationError("昵称太长"))
	}

	if !utils.IsEmail(c.Email) {
		panic(exception.NewValidationError("邮箱不正确"))
	}

	if c.Url != "" && !utils.IsURL(c.Url) {
		panic(exception.NewValidationError("网址不正确"))
	}

	if c.Source == "" {
		panic(exception.NewValidationError("评论内容不能为空"))
	}

	if utf8.RuneCountInString(c.Content) >= 4096 {
		panic(exception.NewValidationError("评论内容太长"))
	}

	if c.Parent > 0 {
		pc := s.GetComment2(c.Parent)
		comment.Root = pc.Root
		if pc.Root == 0 {
			comment.Root = pc.ID
		}
	}

	comment.Content = s.convertCommentMarkdown(c)

	adminEmail := s.cfg.Comment.Email

	if user.IsGuest() {
		notAllowedEmails := s.cfg.Comment.NotAllowedEmails
		if adminEmail != "" {
			notAllowedEmails = append(notAllowedEmails, adminEmail)
		}
		// TODO use regexp to detect equality.
		for _, email := range notAllowedEmails {
			if email != "" && c.Email != "" && strings.EqualFold(email, c.Email) {
				panic(exception.NewValidationError("不能使用此邮箱地址"))
			}
		}
		notAllowedAuthors := s.cfg.Comment.NotAllowedAuthors
		if adminName := s.cfg.Comment.Author; adminName != "" {
			notAllowedAuthors = append(notAllowedAuthors, adminName)
		}
		for _, author := range notAllowedAuthors {
			if author != "" && c.Author != "" && strings.EqualFold(author, string(c.Author)) {
				panic(exception.NewValidationError("不能使用此昵称"))
			}
		}
	}

	s.TxCall(func(txs *Service) error {
		txs.tdb.Model(&comment).MustCreate()
		count := txs.GetAllCommentsCount()
		txs.SetOption("comment_count", count)
		txs.UpdatePostCommentCount(comment.PostID)
		return nil
	})

	s.doCommentNotification(&comment)

	return comment.ToProtocols(adminEmail, user)
}

func (s *Service) convertCommentMarkdown(c *protocols.Comment) string {
	if c.SourceType != "markdown" {
		panic(exception.NewValidationError("仅支持 markdown"))
	}

	var buf bytes.Buffer
	md := goldmark.New(goldmark.WithExtensions(extension.GFM))
	if err := md.Convert([]byte(c.Source), &buf); err != nil {
		panic(exception.NewValidationError("不能转换 markdown"))
	}
	bs := buf.String()
	if strings.Contains(bs, `<!-- raw HTML omitted -->`) {
		panic(exception.NewValidationError(`不能包含 HTML 标签`))
	}
	return bs
}

func (s *Service) DeleteComment(ctx context.Context, commentName int64) {
	cmt := s.GetComment2(commentName)
	s.comments().Where("id=? OR root=?", commentName, commentName).Delete()
	s.UpdatePostCommentCount(cmt.PostID)
}

// SetCommentPostID 把某条顶级评论及其子评论转移到另一篇文章下
func (s *Service) SetCommentPostID(ctx context.Context, in *protocols.SetCommentPostIDRequest) (*protocols.SetCommentPostIDResponse, error) {
	user := s.auth.AuthGRPC(ctx)
	if !user.IsAdmin() {
		panic(403)
	}

	s.TxCall(func(txs *Service) error {
		cmt := s.GetComment2(in.Id)
		if cmt.Root != 0 {
			panic(`不能转移子评论`)
		}
		post := s.GetPostByID(in.PostId)
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
		return nil
	})

	return &protocols.SetCommentPostIDResponse{}, nil
}

func (s *Service) doCommentNotification(c *models.Comment) {
	postTitle := s.GetPostTitle(c.PostID)
	postLink := fmt.Sprintf("%s/%d/", s.HomeURL(), c.PostID)
	adminEmail := s.cfg.Comment.Email
	if adminEmail == "" {
		return
	}

	adminEmail = strings.ToLower(adminEmail)
	commentEmail := strings.ToLower(c.Email)

	if commentEmail != adminEmail {
		data := &comment_notify.AdminData{
			Title:    postTitle,
			Link:     postLink,
			Date:     c.Date,
			Author:   c.Author,
			Content:  c.Source,
			Email:    c.Email,
			HomePage: c.URL,
		}
		if data.Content == "" {
			data.Content = c.Content
		}
		s.cmtntf.NotifyAdmin(data)
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
		email := strings.ToLower(parent.Email)
		if email == adminEmail || email == commentEmail {
			continue
		}
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
		Date:    c.Date,
		Author:  c.Author,
		Content: c.Source,
	}
	if guestData.Content == "" {
		guestData.Content = c.Content
	}

	s.cmtntf.NotifyGuests(&guestData, distinctNames, distinctEmails)
}
