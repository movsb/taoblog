package main

import (
	"bytes"
	"database/sql"
	"fmt"
)

func createSitemap(tx *sql.DB, host string) (string, error) {
	query := `SELECT id FROM posts WHERE type='post' AND status='public' ORDER BY date DESC`
	rows, err := tx.Query(query)
	if err != nil {
		return "", err
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

	sb := bytes.NewBuffer(nil)
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`)
	for _, id := range ids {
		sb.WriteString(fmt.Sprintf("<url><loc>%s/%d/</loc></url>\n", host, id))
	}

	sb.WriteString("</urlset>\n")

	return sb.String(), err
}
