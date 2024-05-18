package admin

import "github.com/movsb/taoblog/cmd/config"

type Option func(a *Admin)

func WithCustomThemes(t *config.ThemeConfig) Option {
	return func(a *Admin) {
		a.customTheme = t
	}
}
