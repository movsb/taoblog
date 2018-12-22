package server

// IServer is implemented by blog server & cache server.
type IServer interface {

	// Comments
	ListComments(in *ListCommentsRequest) *ListCommentsResponse
}
