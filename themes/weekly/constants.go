package weekly

import (
	"regexp"
)

var (
	regexpHome   = regexp.MustCompile(`^/$`)
	regexpBySlug = regexp.MustCompile(`^/(\d+)/$`)
)
