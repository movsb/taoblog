package cache

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taorm"
)

type A struct{}
type B struct{}

type C struct {
	N int
}

func TestTypes(t *testing.T) {
	ti := reflect.TypeFor[A]()
	t.Log(ti.PkgPath())
	t.Log(ti.Name())

	t.Log(typeHash(A{}))
	t.Log(typeHash(B{}))

	db := taorm.NewDB(migration.InitCache(``))
	h := NewFileCache(context.Background(), db)
	c := C{}
	var out []byte
	// 无缓存
	err := h.GetOrLoad(c, time.Second, &out, func() (any, error) {
		return []byte(`data`), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `data` {
		panic(`not data`)
	}
	// 从缓存
	out = []byte{}
	err = h.GetOrLoad(c, time.Second, &out, func() (any, error) {
		return []byte(`data`), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `data` {
		panic(`not data`)
	}
	h.Delete(c)
	c.N = 10
	err = h.GetOrLoad(&c, time.Second, &out, func() (any, error) {
		return []byte(`xxxx`), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `xxxx` {
		panic(`not xxxx`)
	}

	var cs []*C
	h.GetAllKeysFor(&cs)
	if !reflect.DeepEqual(cs, []*C{{N: 10}}) {
		t.Fatal(`not equal`)
	}

	var cs2 []C
	h.GetAllKeysFor(&cs2)
	if !reflect.DeepEqual(cs2, []C{{N: 10}}) {
		t.Fatal(`not equal`)
	}
}

func BenchmarkTypeHash(b *testing.B) {
	for range b.N {
		typeHash(A{})
		typeHash(B{})
	}
}
