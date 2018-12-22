package protocols

import "io"

type GetAvatarRequest struct {
	Query           string
	IfModifiedSince string
	IfNoneMatch     string
	SetStatus       func(statusCode int)
	SetHeader       func(name string, value string)
	W               io.Writer
}
