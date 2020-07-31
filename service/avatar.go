package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/avatar"
)

// GetAvatar ...
func (s *Service) GetAvatar(in *protocols.GetAvatarRequest) {
	p := avatar.Params{
		Headers: make(http.Header),
	}

	if in.IfModifiedSince != "" {
		p.Headers.Add("If-Modified-Since", in.IfModifiedSince)
	}
	if in.IfNoneMatch != "" {
		p.Headers.Add("If-None-Match", in.IfNoneMatch)
	}

	var comment models.Comment
	s.tdb.Select(`email`).Where(`id = ?`, in.CommentID).MustFind(&comment)

	resp, err := avatar.Get(comment.Email, &p)
	if err != nil {
		in.SetStatus(500)
		fmt.Fprint(in.W, err)
		return
	}

	defer resp.Body.Close()

	for name, value := range resp.Header {
		in.SetHeader(name, value[0])
	}

	in.SetStatus(resp.StatusCode)

	io.Copy(in.W, resp.Body)
}
