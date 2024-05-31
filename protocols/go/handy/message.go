package handy

import (
	"io"
)

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

type ListTagsWithCountRequest struct {
}
