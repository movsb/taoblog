package main

import (
	"bytes"
	"html/template"
)

// TODO support Last-Modified

const feedTemplate = `
<rss version="2.0">
	<channel>
		<title>{{.BlogName}}</title>
		<link>{{.Home}}</link>
		<description>{{.Description}}</description>
		{{range .Posts}}
		<item>
			<title>{{.Title}}</title>
			<link>{{.Link}}</link>
			<pubDate>{{.Date}}</pubDate>
			<description>{{.Content}}</description>
		</item>
		{{end}}
	</channel>
</rss>
`

func theFeed(tx Querier) (string, error) {
	buf := bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>`)

	tmpl, err := template.New("feed").Parse(feedTemplate)
	if err != nil {
		return "", err
	}

	posts, err := postmgr.GetPostsForRss(tx)
	if err != nil {
		return "", err
	}

	data := RssData{}
	data.Home = "https://" + optmgr.GetDef(tx, "home", "localhost")
	data.BlogName = optmgr.GetDef(tx, "blog_name", "TaoBlog")
	data.Description = optmgr.GetDef(tx, "desc", "")
	data.Posts = posts

	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
