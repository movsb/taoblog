package github

import "net/http"

func New(secret, reloaderPath string, sendNotify func(content string)) http.Handler {
	return handler(secret, reloaderPath, sendNotify)
}
