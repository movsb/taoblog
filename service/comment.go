package service

import (
	"context"
	"strings"

	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) comments() *taorm.Stmt {
	return s.tdb.Model(models.Comment{}, "comments")
}

// GetComment ...
func (s *ImplServer) GetComment(name int64) *models.Comment {
	var comment models.Comment
	s.comments().Where("id=?", name).Find(&comment)
	return &comment
}

// ListComments ...
func (s *ImplServer) ListComments(in *ListCommentsRequest) []*models.Comment {
	var comments []*models.Comment
	s.comments().Select(in.Fields).Limit(in.Limit).Offset(in.Offset).OrderBy(in.OrderBy).
		WhereIf(in.PostID > 0, "post_id=?", in.PostID).
		WhereIf(in.Ancestor >= 0, "ancestor=?", in.Ancestor).Find(&comments)
	return comments
}

func (s *ImplServer) GetAllCommentsCount() int64 {
	type GetAllCommentsCount_Result struct {
		Count int64
	}
	var result GetAllCommentsCount_Result
	query := `SELECT count(1) as count FROM comments`
	taorm.MustQueryRows(&result, s.db, query)
	return result.Count
}

func (s *ImplServer) CreateComment(ctx context.Context, c *models.Comment) *models.Comment {
	user := s.auth.AuthContext(ctx)
	if user.IsGuest() {
		notAllowedEmails := strings.Split(s.GetDefaultStringOption("not_allowed_emails", ""), ",")
		if adminEmail := s.GetDefaultStringOption("email", ""); adminEmail != "" {
			notAllowedEmails = append(notAllowedEmails, adminEmail)
		}
		// TODO use regexp to detect equality.
		for _, email := range notAllowedEmails {
			if email != "" && c.Email != "" && strings.EqualFold(email, c.Email) {
				panic("不能使用此邮箱地址")
			}
		}
		notAllowedAuthors := strings.Split(s.GetDefaultStringOption("not_allowed_authors", ""), ",")
		if adminName := s.GetDefaultStringOption("nickname", ""); adminName != "" {
			notAllowedAuthors = append(notAllowedAuthors, adminName)
		}
		for _, author := range notAllowedAuthors {
			if author != "" && c.Author != "" && strings.EqualFold(author, string(c.Author)) {
				panic("不能使用此昵称")
			}
		}
	}

	s.tdb.TxCall(func(tx *taorm.DB) {
		tx.Model(c, "comments").Create()
	})

	// TODO wrap in tx
	count := s.GetAllCommentsCount()
	s.SetOption("comment_count", count)
	s.UpdatePostCommentCount(c.PostID)

	return c
}

func (s *ImplServer) DeleteComment(ctx context.Context, commentName int64) {
	cmt := s.GetComment(commentName)
	s.comments().Or("id=?", commentName).Or("ancestor=?", commentName).Delete()
	s.UpdatePostCommentCount(cmt.PostID)
}
