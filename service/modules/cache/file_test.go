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
	val, err := h.GetOrLoad(context.Background(), c, func(ctx context.Context, key FileCacheKey) (value []byte, ttl time.Duration, err error) {
		return []byte(`data`), time.Second * 10, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != `data` {
		panic(`not data`)
	}
	h.Delete(c)
	val, err = h.GetOrLoad(context.Background(), c, func(ctx context.Context, key FileCacheKey) (value []byte, ttl time.Duration, err error) {
		return []byte(`xxxx`), time.Second * 10, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != `xxxx` {
		panic(`not xxxx`)
	}
}

func BenchmarkTypeHash(b *testing.B) {
	for range b.N {
		typeHash(A{})
		typeHash(B{})
	}
}
