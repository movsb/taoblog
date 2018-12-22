package taorm

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test(t *testing.T) {
	var err error
	dataSource := fmt.Sprintf("%[1]s:%[1]s@/%[1]s", "taoblog")
	gdb, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	defer gdb.Close()

	c := struct {
		Metas  string
		_Metas string
	}{}
	if err := QueryRows(&c, gdb, `SELECT * FROM posts`); err != nil {
		panic(err)
	}
	fmt.Println(c)
}
