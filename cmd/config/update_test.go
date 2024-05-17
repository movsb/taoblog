package config

import (
	"encoding/json"
	"fmt"
	"testing"
)

func assert(conditions ...any) {
	if len(conditions)&1 != 0 {
		panic(`not even`)
	}
	for i := 0; i < len(conditions)/2; i++ {
		a, b := conditions[i<<1+0], conditions[i<<1+1]
		if a != b {
			panic(fmt.Sprint(i/2, a, b))
		}
	}
}

func jsonOf(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func TestUpdate(t *testing.T) {
	type _kv struct {
		Key   string
		Value string
	}
	c1, c2 := DefaultConfig(), DefaultConfig()
	u := NewUpdater(&c2)
	for _, s := range []_kv{
		{"database.engine", "mysql"},
		{"database.sqlite", "{}"},
		{`menus[1]`, `{"name":"后台"}`},
	} {
		u.MustApply(s.Key, s.Value)
	}
	assert(
		c1.Database.Engine, `sqlite`,
		c2.Database.Engine, `mysql`,
		c1.Database.SQLite.Path, `taoblog.db`,
		c2.Database.SQLite.Path, "",
		jsonOf(c1.Menus), `[{"Name":"首页","Link":"/","Blank":false,"Items":null},{"Name":"管理后台","Link":"/admin/","Blank":false,"Items":null}]`,
		jsonOf(c2.Menus), `[{"Name":"首页","Link":"/","Blank":false,"Items":null},{"Name":"后台","Link":"","Blank":false,"Items":null}]`,
	)
}
