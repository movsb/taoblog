package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/movsb/taorm/taorm"
)

type Post struct {
	ID            int64
	Date          string
	Modified      string
	Title         string
	Content       string
	Slug          string
	Type          string
	Category      uint `taorm:"name:taxonomy"`
	Status        string
	PageView      uint
	CommentStatus uint
	Comments      uint
	Metas         string
	Source        string
	SourceType    string
}

func main() {
	var err error

	dataSource := fmt.Sprintf("%s:%s@/%s",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
	)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(10)
	defer db.Close()
	tdb := taorm.NewDB(db)
	var posts []Post
	tdb.From("posts").Select("id,title").MustFind(&posts)
	re := regexp.MustCompile(`【(.+)】(.+)`)
	for _, post := range posts {
		post.Title = strings.TrimLeft(post.Title, " ")
		matches := re.FindStringSubmatch(post.Title)
		if len(matches) != 3 {
			continue
		}
		post.Title = fmt.Sprintf(`[%s] %s`, matches[1], matches[2])
		tdb.Model(&post, "posts").MustUpdateMap(map[string]interface{}{
			"title": post.Title,
		})
	}
}
