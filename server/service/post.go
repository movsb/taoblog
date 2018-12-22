package service

import (
	"github.com/movsb/taoblog/modules/server"
	"github.com/movsb/taoblog/server/models"
	"github.com/movsb/taoblog/server/modules/taorm"
)

// GetPost ...
func (s *ImplServer) GetPost(in *server.GetPostRequest) *server.Post {
	query := `SELECT * FROM posts WHERE id = ?`
	var post models.Post
	taorm.MustQueryRows(&post, s.db, query, in.Name)
	return post.Serialize()
}

// ListPosts ...
func (s *ImplServer) ListPosts(in *server.ListPostsRequest) *server.ListPostsResponse {
	query := `SELECT * FROM posts`
	var posts models.Posts
	taorm.MustQueryRows(&posts, s.db, query)
	return &server.ListPostsResponse{
		Posts: posts.Serialize(),
	}
}
