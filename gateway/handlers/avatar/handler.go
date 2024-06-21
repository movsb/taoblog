package avatar

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc/status"
)

func New(auther *auth.Auth, client clients.Client) http.Handler {
	return &_Avatar{
		auther: auther,
		client: client,
	}
}

type _Avatar struct {
	auther *auth.Auth
	client clients.Client
}

// Params ...
type Params struct {
	Headers http.Header
}

func (h *_Avatar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ctx := h.auther.NewContextForRequestAsGateway(r)
	emailRsp, err := h.client.GetCommentEmailById(
		auth.SystemAdminForGateway(r.Context()),
		&proto.GetCommentEmailByIdRequest{
			Id: int32(utils.MustToInt64(r.PathValue(`id`))),
		},
	)
	if err != nil {
		http.Error(w, err.Error(), runtime.HTTPStatusFromCode(status.Code(err)))
		return
	}

	p := Params{
		Headers: make(http.Header),
	}

	for _, name := range []string{
		`If-Modified-Since`,
		`If-None-Match`,
	} {
		if h := r.Header.Get(name); h != "" {
			p.Headers.Add(name, h)
		}
	}

	// TODO 并没有限制获取未公开发表文章的评论。
	rsp, err := github(context.TODO(), emailRsp.Email, &p)
	if err != nil {
		rsp, err = gravatar(context.TODO(), emailRsp.Email, &p)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer rsp.Body.Close()

	// TODO：内部缓存，只正向代理 body。
	for _, k := range knownHeaders {
		if v := rsp.Header.Get(k); v != "" {
			w.Header().Set(k, v)
		}
	}

	// 客户端缓存失效了也可以继续用，后台慢慢刷新就行。
	w.Header().Set(`Cache-Control`, `max-age=604800, stale-while-revalidate=604800`)

	w.WriteHeader(rsp.StatusCode)
	io.Copy(w, rsp.Body)
}

// 不再提供以下字段，官方更新太频繁，意义不大。
// `Expires`,
// `Cache-Control`,
var knownHeaders = []string{
	`Content-Length`,
	`Content-Type`,
	`Last-Modified`,
}
