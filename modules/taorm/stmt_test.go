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

	var comments []*Comment

	tdb.Model(Comment{ID: 1}, "comments").Find(&comments)
	for _, comment := range comments {
		t.Log(comment)
	}
}
