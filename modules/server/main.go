package server

// IServer is implemented by blog server & cache server.
type IServer interface {

	// Comments
	GetComment(in *GetCommentRequest) *Comment
	ListComments(in *ListCommentsRequest) *ListCommentsResponse
}
