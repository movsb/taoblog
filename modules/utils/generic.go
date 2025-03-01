package utils

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	mr "math/rand"
	"net/http"
)

// 搁这套娃🪆🪆🪆？
// P：Prototype
func ChainFuncs[P func(H) H, H any](h H, ps ...P) H {
	for i := len(ps) - 1; i >= 0; i-- {
		h = ps[i](h)
	}
	return h
}

type ServeMuxChain struct {
	*http.ServeMux
}

func (m *ServeMuxChain) Handle(pattern string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) {
	m.ServeMux.Handle(pattern, ChainFuncs(handler, middlewares...))
}

func Must1[A any](a A, e error) A {
	if e != nil {
		panic(e)
	}
	return a
}
func Must(e error) {
	if e != nil {
		panic(e)
	}
}

// Go 语言多少有点儿大病，以至于我需要写这种东西。
// 是谁当初说不需要三元运算符的？我打断他的 🐶 腿。
// https://en.wikipedia.org/wiki/IIf
// https://blog.twofei.com/716/#没有条件运算符
func IIF[Any any](cond bool, first, second Any) Any {
	if cond {
		return first
	}
	return second
}

func RandomString() string {
	b := [4]byte{}
	rand.Read(b[:])
	return fmt.Sprintf(`xx-%x`, b)
}

func ReInitTestRandomSeed() {
	rand.Reader = mr.New(mr.NewSource(0))
}

func DropLast1[First any, Last any](f First, l Last) First {
	return f
}
func KeepLast1[First any, Last any](f First, l Last) Last {
	return l
}

func CatchAsError(err *error) {
	if er := recover(); er != nil {
		if er2, ok := er.(error); ok {
			*err = er2
			return
		}
		*err = fmt.Errorf(`%v`, er)
	}
}

// 基于内存实现的可重复读的 Reader。
// https://blog.twofei.com/1072/
func MemDupReader(r io.Reader) func() io.Reader {
	b := bytes.NewBuffer(nil)
	t := io.TeeReader(r, b)

	return func() io.Reader {
		br := bytes.NewReader(b.Bytes())
		return io.MultiReader(br, t)
	}
}

func Filter[S []E, E any](s S, predicate func(e E) bool) []E {
	r := make([]E, 0, len(s))
	for _, a := range s {
		if predicate(a) {
			r = append(r, a)
		}
	}
	return r
}

func Implements[T any](a any) bool {
	_, ok := a.(T)
	return ok
}

func Map[T any, S []E, E any](s S, mapper func(e E) T) []T {
	t := make([]T, 0, len(s))
	for _, a := range s {
		t = append(t, mapper(a))
	}
	return t
}

// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

type PluginStorage interface {
	Set(key string, value string) error
	Get(key string) (string, error)
}

type InMemoryStorage struct {
	m map[string]string
}

func (s *InMemoryStorage) Set(key string, value string) error {
	s.m[key] = value
	return nil
}

func (s *InMemoryStorage) Get(key string) (string, error) {
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return ``, sql.ErrNoRows
}

func NewInMemoryStorage() PluginStorage {
	return &InMemoryStorage{
		m: map[string]string{},
	}
}
