package config_test

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/movsb/taoblog/cmd/config"
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
	t.SkipNow()
	type _kv struct {
		Key   string
		Value string
	}
	c1, c2 := config.DefaultConfig(), config.DefaultConfig()
	u := config.NewUpdater(c2)
	for _, s := range []_kv{
		{`menus[1]`, `{"name":"后台"}`},
	} {
		u.MustApply(s.Key, s.Value, func(path, value string) {})
	}
	assert(
		jsonOf(c1.Menus), `[{"Name":"首页","Link":"/","Blank":false,"Items":null},{"Name":"管理后台","Link":"/admin/","Blank":false,"Items":null}]`,
		jsonOf(c2.Menus), `[{"Name":"首页","Link":"/","Blank":false,"Items":null},{"Name":"后台","Link":"","Blank":false,"Items":null}]`,
	)
}

type Config struct {
	A A `json:"a" yaml:"a"`
	B B `json:"b" yaml:"b"`
}

type A struct {
	AA int `json:"aa" yaml:"aa"`
}

func (A) CanSave() {}

type B struct {
	C C `json:"c" yaml:"c"`
}

type C struct {
	CC string `json:"cc" yaml:"cc"`
}

func (C) CanSave() {}

func (C) AfterSet(paths config.Segments, obj any) {
	log.Println(`AfterSet:`, paths, obj)
}

func TestSaver(t *testing.T) {
	var c Config
	updater := config.NewUpdater(&c)

	var c2 Config
	updater2 := config.NewUpdater(&c2)

	updater.MustApply("a.aa", "123", func(path, value string) {
		t.Logf("save: %s: %s", path, value)
		updater2.MustApply(path, value, func(path, value string) {})
	})
	updater.MustApply("b.c.cc", "123", func(path, value string) {
		t.Logf("save: %s: %s", path, value)
		updater2.MustApply(path, value, func(path, value string) {})
	})
	if !reflect.DeepEqual(c, c2) {
		t.Errorf(`不相等。`)
	}
	updater2.EachSaver(func(path string, obj any) {
		log.Println(`将会保存：`, path, obj)
	})
}

type Set struct {
	A int `yaml:"a"`
}

func (Set) CanSave() {}
func (s *Set) BeforeSet(paths config.Segments, obj any) error {
	if paths.At(0).Key != `a` {
		panic(`expect a`)
	}
	n, ok := obj.(int)
	if !ok || n != 123 {
		panic(`expect 123`)
	}
	return nil
}
func (s *Set) AfterSet(paths config.Segments, obj any) {
	if paths.At(0).Key != `a` {
		panic(`expect a`)
	}
	n, ok := obj.(int)
	if !ok || n != 123 {
		panic(`expect 123`)
	}
}

func TestSaver2(t *testing.T) {
	var a Set
	updater := config.NewUpdater(&a)
	updater.MustApply(`a`, `123`, func(path, value string) {
		t.Logf(`保存：%s: %s`, path, value)
	})
}
