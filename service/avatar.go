package service

import (
	"io"
	"net/http"

	"github.com/movsb/taoblog/protocols"
)

const (
	gGrAvatarHost = "https://www.gravatar.com/avatar/"
)

func (s *ImplServer) GetAvatar(in *protocols.GetAvatarRequest) {
	u := gGrAvatarHost + in.Query
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}

	if in.IfModifiedSince != "" {
		req.Header.Set("If-Modified-Since", in.IfModifiedSince)
	}
	if in.IfNoneMatch != "" {
		req.Header.Set("If-None-Match", in.IfNoneMatch)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	in.SetStatus(resp.StatusCode)

	for name, value := range resp.Header {
		in.SetHeader(name, value[0])
	}

	io.Copy(in.W, resp.Body)
}
