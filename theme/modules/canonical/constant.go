package canonical

import "regexp"

var (
	regexpHome   = regexp.MustCompile(`^/$`)
	regexpByID   = regexp.MustCompile(`^/(\d+)(/?)$`)
	regexpFile   = regexp.MustCompile(`^/(\d+)/(.+)$`)
	regexpByTags = regexp.MustCompile(`^/tags/(.*)$`)
	regexpByPage = regexp.MustCompile(`^((/[0-9a-zA-Z\-_]+)*)/([0-9a-zA-Z\-_]+)$`)
)

var nonCategoryNames = map[string]bool{
	"/admin/":   true,
	"/scripts/": true,
	"/images/":  true,
}
