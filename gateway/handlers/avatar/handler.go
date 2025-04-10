package avatar

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"
)

type ResolveFunc func(id uint32) (email string, mod time.Time, file io.ReadSeekCloser)

// userAvatars: 通过以邮箱作为文件名查询头像。
func New(task *Task, resolve ResolveFunc) *_Avatar {
	return &_Avatar{
		task:    task,
		resolve: resolve,
	}
}

type _Avatar struct {
	task    *Task
	resolve ResolveFunc
}

func (h *_Avatar) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseUint(r.PathValue(`id`), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 客户端缓存失效了也可以继续用，后台慢慢刷新就行。
		// 如果失败，会被 ServeFileFS 自动删除。
		w.Header().Set(`Cache-Control`, `max-age=604800, stale-while-revalidate=604800`)

		email, mod, user := h.resolve(uint32(id))
		if email != `` {
			l, c, found := h.task.Get(email)
			if found {
				http.ServeContent(w, r, `avatar`, l, bytes.NewReader(c))
				return
			}
		} else if user != nil {
			defer user.Close()
			http.ServeContent(w, r, `avatar`, mod, user)
			return
		}
		http.NotFound(w, r)
	})
}
