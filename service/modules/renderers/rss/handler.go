package rss

import (
	"net/http"
	"time"

	"github.com/movsb/taoblog/modules/auth"
)

type Handler struct {
	t *Task
}

func NewHandler(t *Task) http.Handler {
	h := &Handler{t: t}
	mux := http.NewServeMux()
	mux.HandleFunc(`/open`, h.open)
	return auth.RequireLogin(mux)
}

func (h *Handler) open(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Query().Get(`url`)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	go h.markRead(u)
}

func (h *Handler) markRead(url string) {
	h.t.lock.Lock()
	defer h.t.lock.Unlock()

	for pid, sites := range h.t.posts {
		for _, posts := range sites {
			for _, post := range posts {
				if post.PostURL == url {
					post.ReadAt = time.Now()
					h.t.saveDebouncer.Enter()
					h.t.invalidate(pid)
					break
				}
			}
		}
	}
}
