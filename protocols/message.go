package protocols

import "io"

type GetAvatarRequest struct {
	Ephemeral       int
	IfModifiedSince string
	IfNoneMatch     string
	SetStatus       func(statusCode int)
	SetHeader       func(name string, value string)
	W               io.Writer
}

type ListLatestPostsRequest struct {
}

type ListPostsRequest struct {
	Fields  string
	Limit   int64
	OrderBy string

	WithContent bool
}

type ListTagsWithCountRequest struct {
}
