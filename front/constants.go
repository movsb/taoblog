package front

import (
	"regexp"
)

var (
	regexpHome   = regexp.MustCompile(`^/$`)
	regexpByID   = regexp.MustCompile(`/(\d+)/$`)
	regexpFile   = regexp.MustCompile(`/(\d+)/(.+)$`)
	regexpBySlug = regexp.MustCompile(`^/(.+)/([^/]+)\.html$`)
	regexpByTags = regexp.MustCompile(`^/tags/(.*)$`)
	regexpByPage = regexp.MustCompile(`^((/[0-9a-zA-Z\-_]+)*)/([0-9a-zA-Z\-_]+)$`)
)

var nonCategoryNames = map[string]bool{
	"/admin/":    true,
	"/emotions/": true,
	"/scripts/":  true,
	"/images/":   true,
	"/sass/":     true,
	"/tags/":     true,
	"/plugins/":  true,
	"/files/":    true,
}
