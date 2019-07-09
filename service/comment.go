package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/movsb/taoblog/service/modules/comment_notify"

	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm/taorm"
)

func (s *Service) comments() *taorm.Stmt {
	return s.tdb.Model(models.Comment{}, "comments")
}

// GetComment ...
func (s *Service) GetComment(name int64) *models.Comment {
	var comment models.Comment
	s.comments().Where("id=?", name).MustFind(&comment)
	return &comment
}

func (s *Service) ListComments(ctx context.Context, in *protocols.ListCommentsRequest) []*protocols.Comment {
	user := s.auth.AuthContext(ctx)
	adminEmail := s.GetStringOption("email")

	var parentProtocolComments []*protocols.Comment
	{
		var parents models.Comments
		// TODO ensure that fields must include root etc to be used later.
		stmt := s.tdb.From("comments").Select(in.Fields)
		stmt.WhereIf(in.Mode == protocols.ListCommentsModeTree, "root = 0")
		stmt.WhereIf(in.PostID > 0, "post_id=?", in.PostID)
		// limit & offset apply to parent comments only
		stmt.Limit(in.Limit).Offset(in.Offset).OrderBy(in.OrderBy)
		if user.IsGuest() {
			stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
		}
		stmt.MustFind(&parents)
		parentProtocolComments = parents.ToProtocols(adminEmail, user)
	}

	needChildren := in.Mode == protocols.ListCommentsModeTree && len(parentProtocolComments) > 0

	if needChildren {
		var childrenProtocolComments []*protocols.Comment
		{
			parentIDs := make([]int64, 0, len(parentProtocolComments))
			for _, parent := range parentProtocolComments {
				parentIDs = append(parentIDs, parent.ID)
			}
			var children models.Comments
			stmt := s.tdb.From("comments").Select(in.Fields)
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
				parent.Children = childrenMap[parent.ID]
			}
		}
	}

	return parentProtocolComments
}

func (s *Service) GetAllCommentsCount() int64 {
	type GetAllCommentsCount_Result struct {
		Count int64
	}
	var result GetAllCommentsCount_Result
	s.tdb.Model(models.Comment{}, "comments").Select("count(1) as count").Find(&result)
	return result.Count
}

func (s *Service) CreateComment(ctx context.Context, c *protocols.Comment) *protocols.Comment {
	user := s.auth.AuthContext(ctx)

	comment := models.Comment{
		PostID:  c.PostID,
		Parent:  c.Parent,
		Author:  c.Author,
		Email:   c.Email,
		URL:     c.URL,
		IP:      c.IP,
		Date:    c.Date,
		Content: c.Content,
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

	if c.URL != "" && !utils.IsURL(c.URL) {
		panic(exception.NewValidationError("网址不正确"))
	}

	if c.Content == "" {
		panic(exception.NewValidationError("评论内容不能为空"))
	}

	if utf8.RuneCountInString(c.Content) >= 4096 {
		panic(exception.NewValidationError("评论内容太长"))
	}

	if c.Parent > 0 {
		pc := s.GetComment(c.Parent)
		comment.Root = pc.Root
		if pc.Root == 0 {
			comment.Root = pc.ID
		}
	}

	adminEmail := s.GetDefaultStringOption("email", "")

	if user.IsGuest() {
		notAllowedEmails := strings.Split(s.GetDefaultStringOption("not_allowed_emails", ""), ",")
		if adminEmail != "" {
			notAllowedEmails = append(notAllowedEmails, adminEmail)
		}
		// TODO use regexp to detect equality.
		for _, email := range notAllowedEmails {
			if email != "" && c.Email != "" && strings.EqualFold(email, c.Email) {
				panic(exception.NewValidationError("不能使用此邮箱地址"))
			}
		}
		notAllowedAuthors := strings.Split(s.GetDefaultStringOption("not_allowed_authors", ""), ",")
		if adminName := s.GetDefaultStringOption("author", ""); adminName != "" {
			notAllowedAuthors = append(notAllowedAuthors, adminName)
		}
		for _, author := range notAllowedAuthors {
			if author != "" && c.Author != "" && strings.EqualFold(author, string(c.Author)) {
				panic(exception.NewValidationError("不能使用此昵称"))
			}
		}
	}

	s.TxCall(func(txs *Service) error {
		txs.tdb.Model(&comment, "comments").MustCreate()
		count := txs.GetAllCommentsCount()
		txs.SetOption("comment_count", count)
		txs.UpdatePostCommentCount(comment.PostID)
		return nil
	})

	s.doCommentNotification(&comment)

	return comment.ToProtocols(adminEmail, user)
}

func (s *Service) DeleteComment(ctx context.Context, commentName int64) {
	cmt := s.GetComment(commentName)
	s.comments().Or("id=?", commentName).Or("root=?", commentName).Delete()
	s.UpdatePostCommentCount(cmt.PostID)
}

func (s *Service) doCommentNotification(c *models.Comment) {
	postTitle := s.GetPostTitle(c.PostID)
	postLink := fmt.Sprintf("%s/%d/", s.HomeURL(), c.PostID)
	adminEmail := s.GetDefaultStringOption("email", "")
	if adminEmail == "" {
		return
	}

	adminEmail = strings.ToLower(adminEmail)
	commentEmail := strings.ToLower(c.Email)

	if commentEmail != adminEmail {
		s.cmtntf.NotifyAdmin(&comment_notify.AdminData{
			Title:    postTitle,
			Link:     postLink,
			Date:     c.Date,
			Author:   c.Author,
			Content:  c.Content,
			Email:    c.Email,
			HomePage: c.URL,
		})
	}

	var parents []models.Comment

	for parentID := c.Parent; parentID > 0; {
		var parent models.Comment
		s.tdb.From("comments").
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
		Content: c.Content,
	}

	s.cmtntf.NotifyGuests(&guestData, distinctNames, distinctEmails)
}
