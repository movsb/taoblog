package taorm

import (
	"testing"
)

func TestSQL(t *testing.T) {
	db := dbConn()
	defer db.Close()
	type Post struct {
		Title string
	}
	var ps []Post
	_ = ps
	tdb := NewDB(db)

	mustEqual := func(s1, s2 string) {
		if s1 != s2 {
			t.Logf("\n%s\n%s\n", s1, s2)
			t.Fatal()
		}
	}
	mustEqual(
		tdb.From("comments").Select("content").InnerJoin("posts", "posts.id = comments.post_id").FindSQL(),
		`SELECT comments.content FROM comments INNER JOIN posts ON posts.id = comments.post_id`,
	)
	mustEqual(
		tdb.From("posts").FindSQL(),
		`SELECT * FROM posts`,
	)
	mustEqual(
		tdb.From("posts").Select("id").FindSQL(),
		`SELECT id FROM posts`,
	)
	mustEqual(
		tdb.From("comments").InnerJoin("posts", "").FindSQL(),
		`SELECT comments.* FROM comments INNER JOIN posts`,
	)
	mustEqual(
		tdb.From("comments").Select("*").InnerJoin("posts", "").FindSQL(),
		`SELECT * FROM comments INNER JOIN posts`,
	)
	mustEqual(
		tdb.From("comments").InnerJoin("posts", "comments.post_id=posts.id").FindSQL(),
		`SELECT comments.* FROM comments INNER JOIN posts ON comments.post_id=posts.id`,
	)
	mustEqual(
		tdb.From("comments").UpdateSQL(map[string]interface{}{
			"name": "value",
		}),
		"UPDATE comments SET name=value",
	)
	mustEqual(
		tdb.From("comments").UpdateSQL(map[string]interface{}{
			"name2": Expr("name2+1"),
		}),
		"UPDATE comments SET name2=name2+1",
	)
}
