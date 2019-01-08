package taorm

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	db := dbConn()
	defer db.Close()
	type Post struct {
		Title string
	}
	var ps []Post
	tdb := NewDB(db)
	tdb.From("posts").Select("title").Find(&ps)
	fmt.Println(ps)
}
