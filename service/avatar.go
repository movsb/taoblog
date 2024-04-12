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

	// 删除可能有隐私的头部字段。
	// TODO：内部缓存，只正向代理 body。
	for k := range knownHeaders {
		if v := resp.Header.Get(k); v != "" {
			in.SetHeader(k, v)
		}
	}

	in.SetStatus(resp.StatusCode)

	io.Copy(in.W, resp.Body)
}

var knownHeaders = map[string]bool{
	`Content-Length`: true,
	`Content-Type`:   true,
	`Last-Modified`:  true,
	`Expires`:        true,
	`Cache-Control`:  true,
}
