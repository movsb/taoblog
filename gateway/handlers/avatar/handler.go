package avatar

import (
	"fmt"
	"net/http"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service"
)

func New(task *Task, impl service.ToBeImplementedByRpc) http.Handler {
	return &_Avatar{
		task: task,
		impl: impl,
	}
}

type _Avatar struct {
	task *Task
	impl service.ToBeImplementedByRpc
}

func (h *_Avatar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := h.impl.GetCommentEmailById(int(utils.MustToInt64(r.PathValue(`id`))))
	if email == "" {
		http.Error(w, `找不到对应的邮箱。`, http.StatusNotFound)
		return
	}

	l, ct, c, found := h.task.Get(email)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rl := r.Header.Get(`If-Modified-Since`)
	if rl != `` && rl == l {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// 客户端缓存失效了也可以继续用，后台慢慢刷新就行。
	w.Header().Set(`Cache-Control`, `max-age=604800, stale-while-revalidate=604800`)
	if l != `` {
		w.Header().Set(`Last-Modified`, l)
	}
	if ct != `` {
		w.Header().Set(`Content-Type`, ct)
	}
	w.Header().Set(`Content-Length`, fmt.Sprint(len(c)))
	w.Write(c)
}
