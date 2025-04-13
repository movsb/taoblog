package cache

import (
	"context"
	"reflect"
	"testing"
	"time"
)

type A struct{}
type B struct{}

type C struct {
	N int
}

func (c C) CacheKey() []byte {
	return []byte(`CC`)
}

func TestTypes(t *testing.T) {
	ti := reflect.TypeFor[A]()
	t.Log(ti.PkgPath())
	t.Log(ti.Name())

	t.Log(typeHash(A{}))
	t.Log(typeHash(B{}))

	h := NewFileCache(context.Background(), ``)
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
	err = h.GetOrLoad(c, time.Second, &out, func() (any, error) {
		return []byte(`xxxx`), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `xxxx` {
		panic(`not xxxx`)
	}
}

func BenchmarkTypeHash(b *testing.B) {
	for range b.N {
		typeHash(A{})
		typeHash(B{})
	}
}
