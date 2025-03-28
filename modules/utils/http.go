package utils

import "net/http"

// 类似 RSS 这种总是应该只输出公开文章，完全不用管当前是否登录。
func StripCredentialsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(`Cookie`)
		r.Header.Del(`Authorization`)
		h.ServeHTTP(w, r)
	})
}

type HTTPMux interface {
	Handle(pattern string, handler http.Handler)
}
