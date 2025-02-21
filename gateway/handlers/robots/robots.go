package robots

import (
	"net/http"
	"strings"
	"time"
)

type Robots struct {
	f *File
}

func NewDefaults(sitemapFullURL string) *Robots {
	return &Robots{
		f: &File{
			Groups: []RuleGroup{
				{
					UserAgents: []string{
						`*`,
					},
					Disallows: []string{
						`/admin/`,
						`/v3/`,
						`/search`,
					},
				},
			},
			Sitemap: sitemapFullURL,
		},
	}
}

var now = time.Now()

func (h *Robots) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s := strings.NewReader(h.f.String())
	http.ServeContent(w, r, `robots.txt`, now, s)
}
