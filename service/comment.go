package service

import (
	"github.com/movsb/taoblog/modules/sql_helpers"
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
	seldb := sql_helpers.NewSelect().From("comments", "").Select("*")
	if in.Limit > 0 {
		seldb.Limit(in.Limit)
	}
	if in.OrderBy != "" {
		seldb.OrderBy(in.OrderBy)
	}
	if in.Parent > 0 {
		seldb.Where("post_id = ?", in.Parent)
	}
	query, args := seldb.SQL()
	var comments models.Comments
	taorm.MustQueryRows(&comments, s.db, query, args...)
	return &protocols.ListCommentsResponse{
		Comments: comments.Serialize(),
	}
}
