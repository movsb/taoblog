package avatar

import (
	"io/fs"
	"net/http"
	"strconv"
)

// userAvatars: 通过以邮箱作为文件名查询头像。
func New(task *Task, userAvatars fs.FS) *_Avatar {
	return &_Avatar{
		task:        task,
		userAvatars: userAvatars,
	}
}

type _Avatar struct {
	task        *Task
	userAvatars fs.FS
}

func (h *_Avatar) Ephemeral() http.Handler {
	return http.HandlerFunc(serve(h.task.FS()))
}
func (h *_Avatar) UserID() http.Handler {
	return http.HandlerFunc(serve(h.userAvatars))
}

func serve(fs fs.FS) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := strconv.Atoi(r.PathValue(`id`)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 客户端缓存失效了也可以继续用，后台慢慢刷新就行。
		// 如果失败，会被 ServeFileFS 自动删除。
		w.Header().Set(`Cache-Control`, `max-age=604800, stale-while-revalidate=604800`)

		http.ServeFileFS(w, r, fs, r.PathValue(`id`))
	}
}
