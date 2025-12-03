package admin

import (
	"net/http"

	"github.com/movsb/taoblog/cmd/config"
)

type Option func(a *Admin)

func WithCustomThemes(t *config.ThemeConfig) Option {
	return func(a *Admin) {
		a.customTheme = t
	}
}

func WithWebAuthnHandler(handler func() http.Handler) Option {
	return func(a *Admin) {
		a.webAuthnHandler = handler
	}
}
