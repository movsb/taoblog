package service

import (
	"github.com/movsb/taoblog/modules/server"
	"github.com/movsb/taoblog/server/models"
	"github.com/movsb/taoblog/server/modules/taorm"
)

// ListComments ...
func (s *ImplServer) ListComments(in *server.ListCommentsRequest) *server.ListCommentsResponse {
	query := `SELECT * FROM COMMENTS`
	var comments models.Comments
	taorm.MustQueryRows(&comments, s.db, query)
	return &server.ListCommentsResponse{
		Comments: comments.Serialize(),
	}
}
