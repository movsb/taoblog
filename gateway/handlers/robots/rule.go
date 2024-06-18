package robots

import (
	"fmt"
	"strings"
)

// https://developers.google.com/search/docs/crawling-indexing/robots/create-robots-txt#create_rules
type RuleGroup struct {
	UserAgents []string

	Disallows []string

	// 写在后面，可以解除被 Disallow 禁止的内容。
	Allows []string
}

type File struct {
	Groups []RuleGroup

	// The sitemap URL must be a fully-qualified URL.
	Sitemap string
}

func (f *File) String() string {
	buf := &strings.Builder{}

	for i, g := range f.Groups {
		if i > 0 {
			fmt.Fprintln(buf)
		}
		if len(g.UserAgents) <= 0 || len(g.Allows)+len(g.Disallows) <= 0 {
			continue
		}
		for _, u := range g.UserAgents {
			fmt.Fprintf(buf, "User-agent: %s\n", u)
		}
		for _, a := range g.Allows {
			fmt.Fprintf(buf, "Allow: %s\n", a)
		}
		for _, a := range g.Disallows {
			fmt.Fprintf(buf, "Disallow: %s\n", a)
		}
	}

	if f.Sitemap != "" {
		if buf.Len() > 0 {
			buf.WriteRune('\n')
		}
		fmt.Fprintf(buf, "Sitemap: %s\n", f.Sitemap)
	}

	return buf.String()
}
