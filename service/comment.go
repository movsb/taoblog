package service

import (
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
	s.comments().Select(in.Fields).Limit(in.Limit).OrderBy(in.OrderBy).
		WhereIf(in.Parent > 0, "post_id=?", in.Parent).Find(&comments)
	return comments
}
