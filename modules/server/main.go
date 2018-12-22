package server

// IServer is implemented by blog server & cache server.
type IServer interface {
	// Posts
	GetPost(in *GetPostRequest) *Post
	ListPosts(in *ListPostsRequest) *ListPostsResponse

	// Comments
	GetComment(in *GetCommentRequest) *Comment
	ListComments(in *ListCommentsRequest) *ListCommentsResponse
}
