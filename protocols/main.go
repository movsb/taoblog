package protocols

// IServer is implemented by blog server & cache server.
type IServer interface {
	// Auth
	Auth(in *AuthRequest) *AuthResponse
	AuthLogin(in *AuthLoginRequest) *AuthLoginResponse

	// Posts
	GetPost(in *GetPostRequest) *Post
	ListPosts(in *ListPostsRequest) *ListPostsResponse
	GetPostByID(in *GetPostByIDRequest) *Post
	GetPostBySlug(in *GetPostBySlugRequest) *Post
	GetPostByPage(in *GetPostByPageRequest) *Post
	GetLatestPosts(in *GetLatestPostsRequest) *GetLatestPostsResponse
	GetRelatedPosts(in *GetRelatedPostsRequest) *GetRelatedPostsResponse
	IncrementPostView(in *IncrementPostViewRequest) *IncrementPostViewResponse
	GetPostTitle(ID int64) string
	GetPostTags(ID int64) []string

	// Comments
	GetComment(in *GetCommentRequest) *Comment
	ListComments(in *ListCommentsRequest) *ListCommentsResponse

	// Options
	GetOption(in *GetOptionRequest) *Option
	ListOptions(in *ListOptionsRequest) *ListOptionsResponse

	// Tags
	ListTagsWithCount(in *ListTagsWithCountRequest) *ListTagsWithCountResponse

	// RSS
	GetRss(in *GetRssRequest) *Rss

	// Avatar
	GetAvatar(in *GetAvatarRequest) *Empty

	// Backup
	GetBackup(in *GetBackupRequest) *Empty
}
