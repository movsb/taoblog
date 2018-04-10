package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
)

const sitemapTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<urlset>
{{range .}}<url><loc>/{{.}}</loc></url>
{{end}}
</urlset>
`

func makeSitemap(db *sql.DB) string {
	query := `SELECT id FROM posts WHERE type='post' ORDER BY date DESC`
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Sprint(err)
	}

	defer rows.Close()

	ids := make([]int, 0)

	for rows.Next() {
		id := 0
		if rows.Scan(&id) != nil {
			continue
		}
		ids = append(ids, id)
	}

	tmpl := template.New("sitemap")
	tmpl, err = tmpl.Parse(sitemapTemplate)
	if err != nil {
		return ""
	}

	str := bytes.NewBuffer(nil)
	tmpl.Execute(str, ids)

	return str.String()
}
