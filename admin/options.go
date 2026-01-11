package admin

import (
	"net/http"
)

type Option func(a *Admin)

func WithWebAuthnHandler(handler func() http.Handler) Option {
	return func(a *Admin) {
		a.webAuthnHandler = handler
	}
}
