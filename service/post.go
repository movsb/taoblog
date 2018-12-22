package service

import (
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

// GetPost ...
func (s *ImplServer) GetPost(in *protocols.GetPostRequest) *protocols.Post {
	query := `SELECT * FROM posts WHERE id = ?`
	var post models.Post
	taorm.MustQueryRows(&post, s.db, query, in.Name)
	return post.Serialize()
}

// ListPosts ...
func (s *ImplServer) ListPosts(in *protocols.ListPostsRequest) *protocols.ListPostsResponse {
	query := `SELECT * FROM posts`
	var posts models.Posts
	taorm.MustQueryRows(&posts, s.db, query)
	return &protocols.ListPostsResponse{
		Posts: posts.Serialize(),
	}
}
