package service

import (
	"github.com/movsb/taoblog/modules/server"
	"github.com/movsb/taoblog/server/models"
	"github.com/movsb/taoblog/server/modules/taorm"
)

// GetComment ...
func (s *ImplServer) GetComment(in *server.GetCommentRequest) *server.Comment {
	query := `SELECT * FROM comments WHERE id = ?`
	var comment models.Comment
	taorm.MustQueryRows(&comment, s.db, query, in.Name)
	return comment.Serialize()
}

// ListComments ...
func (s *ImplServer) ListComments(in *server.ListCommentsRequest) *server.ListCommentsResponse {
	query := `SELECT * FROM comments`
	var comments models.Comments
	taorm.MustQueryRows(&comments, s.db, query)
	return &server.ListCommentsResponse{
		Comments: comments.Serialize(),
	}
}
