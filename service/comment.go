package service

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/movsb/taoblog/service/modules/comment_notify"

	"github.com/movsb/taoblog/exception"
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) comments() *taorm.Stmt {
	return s.tdb.Model(models.Comment{}, "comments")
}

// GetComment ...
func (s *ImplServer) GetComment(name int64) *models.Comment {
	var comment models.Comment
	s.comments().Where("id=?", name).MustFind(&comment)
	return &comment
}

// ListComments ...
func (s *ImplServer) ListComments(in *protocols.ListCommentsRequest) []*models.Comment {
	var comments []*models.Comment
	s.comments().Select(in.Fields).Limit(in.Limit).Offset(in.Offset).OrderBy(in.OrderBy).
		WhereIf(in.PostID > 0, "post_id=?", in.PostID).
		WhereIf(in.Ancestor >= 0, "ancestor=?", in.Ancestor).MustFind(&comments)
	return comments
}

func (s *ImplServer) GetAllCommentsCount() int64 {
	type GetAllCommentsCount_Result struct {
		Count int64
	}
	var result GetAllCommentsCount_Result
	s.tdb.Model(models.Comment{}, "comments").Select("count(1) as count").Find(&result)
	return result.Count
}

func (s *ImplServer) CreateComment(ctx context.Context, c *models.Comment) *models.Comment {
	user := s.auth.AuthContext(ctx)

	if c.ID != 0 {
		panic(exception.NewValidationError("评论ID必须为0"))
	}

	if c.Ancestor != 0 {
		panic(exception.NewValidationError("不能指定祖先ID"))
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
		c.Ancestor = pc.Ancestor
		if pc.Ancestor == 0 {
			c.Ancestor = pc.ID
		}
	}

	if user.IsGuest() {
		notAllowedEmails := strings.Split(s.GetDefaultStringOption("not_allowed_emails", ""), ",")
		if adminEmail := s.GetDefaultStringOption("email", ""); adminEmail != "" {
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

	s.TxCall(func(txs *ImplServer) error {
		txs.tdb.Model(c, "comments").Create()
		count := txs.GetAllCommentsCount()
		txs.SetOption("comment_count", count)
		txs.UpdatePostCommentCount(c.PostID)
		return nil
	})

	s.doCommentNotification(c)

	return c
}

func (s *ImplServer) DeleteComment(ctx context.Context, commentName int64) {
	cmt := s.GetComment(commentName)
	s.comments().Or("id=?", commentName).Or("ancestor=?", commentName).Delete()
	s.UpdatePostCommentCount(cmt.PostID)
}

func (s *ImplServer) doCommentNotification(c *models.Comment) {
	home := s.GetDefaultStringOption("home", "localhost")
	postTitle := s.GetPostTitle(c.PostID)
	postLink := fmt.Sprintf("https://%s/%d/", home, c.PostID)
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
