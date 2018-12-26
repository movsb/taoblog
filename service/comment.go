package service

import (
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/service/models"
)

// GetComment ...
func (s *ImplServer) GetComment(name int64) *models.Comment {
	query := `SELECT * FROM comments WHERE id = ?`
	var comment models.Comment
	taorm.MustQueryRows(&comment, s.db, query, name)
	return &comment
}

// ListComments ...
func (s *ImplServer) ListComments(in *ListCommentsRequest) []*models.Comment {
	query := `SELECT * FROM comments`
	var comments []*models.Comment
	taorm.MustQueryRows(&comments, s.db, query)
	return comments
}
