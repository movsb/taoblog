package service

import (
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

// GetComment ...
func (s *ImplServer) GetComment(in *protocols.GetCommentRequest) *protocols.Comment {
	query := `SELECT * FROM comments WHERE id = ?`
	var comment models.Comment
	taorm.MustQueryRows(&comment, s.db, query, in.Name)
	return comment.Serialize()
}

// ListComments ...
func (s *ImplServer) ListComments(in *protocols.ListCommentsRequest) *protocols.ListCommentsResponse {
	query := `SELECT * FROM comments`
	var comments models.Comments
	taorm.MustQueryRows(&comments, s.db, query)
	return &protocols.ListCommentsResponse{
		Comments: comments.Serialize(),
	}
}
