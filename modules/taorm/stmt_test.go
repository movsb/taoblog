package taorm

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestStmt(t *testing.T) {
	var err error
	dataSource := fmt.Sprintf("%[1]s:%[1]s@/%[1]s", "taoblog")
	gdb, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	defer gdb.Close()

	type Comment struct {
		ID       int64
		Parent   int64
		Ancestor int64
		PostID   int64
		Author   string
		Email    string
		URL      string
		IP       string
		Date     string
		Content  string
	}

	tdb := DB{db: gdb}
	_ = tdb

	var comments []*Comment
	out := func() {
		for _, comment := range comments {
			t.Log(comment)
		}
	}
	_ = out

	tdb.Model(Comment{}, "comments").Where("id=?", 100).MustFind(&comments)
	//out()

	//tdb.Model(Comment{}, "comments_copy").UpdateMap(map[string]interface{}{
	//	"author": "桃子",
	//})

	//tdb.Model(Comment{ID: 10000000}, "comments_copy").Limit(1).Delete()

	tdb.Model(comments[0], "comments_copy").Create()
	fmt.Println("new id:", comments[0].ID)
}
