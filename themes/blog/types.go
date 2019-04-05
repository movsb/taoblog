package blog

import (
	"github.com/movsb/taoblog/auth"
)

// TemplateCommon is used by all templates for common data.
type TemplateCommon struct {
	User *auth.User
}

func (t *TemplateCommon) IsAdmin() bool {
	return t.User != nil && !t.User.IsGuest()
}
