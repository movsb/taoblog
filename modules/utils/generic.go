package utils

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	mr "math/rand"
	"net/http"
	"strconv"
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
func Must2[A any, B any](a A, b B, e error) (A, B) {
	if e != nil {
		panic(e)
	}
	return a, b
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
func DropLast2[First any, Second any, Last any](f First, s Second, l Last) (First, Second) {
	return f, s
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

// 对于 Get*Default 方法来说，只会当错误为 key 不存在时会返回默认值。
// 其它时候照常报错，以避免真实错误被隐藏。
type PluginStorage interface {
	SetString(key string, value string) error
	GetString(key string) (string, error)
	GetStringDefault(key string, def string) (string, error)
	SetInteger(key string, value int64) error
	GetInteger(key string) (int64, error)
	GetIntegerDefault(key string, def int64) (int64, error)
	Range(func(key string))
}

type InMemoryStorage struct {
	m map[string]string
}

func (s *InMemoryStorage) SetString(key string, value string) error {
	s.m[key] = value
	return nil
}

func (s *InMemoryStorage) SetInteger(key string, i int64) error {
	s.m[key] = fmt.Sprint(i)
	return nil
}

func (s *InMemoryStorage) GetString(key string) (string, error) {
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return ``, sql.ErrNoRows
}

func (s *InMemoryStorage) GetStringDefault(key string, def string) (string, error) {
	if v, ok := s.m[key]; ok {
		return v, nil
	}
	return def, nil
}
func (s *InMemoryStorage) GetInteger(key string) (int64, error) {
	if v, ok := s.m[key]; ok {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, nil
		} else {
			return 0, err
		}
	}
	return 0, sql.ErrNoRows
}
func (s *InMemoryStorage) GetIntegerDefault(key string, def int64) (int64, error) {
	if v, ok := s.m[key]; ok {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i, nil
		} else {
			return 0, err
		}
	}
	return def, nil
}

func (s *InMemoryStorage) Range(iter func(key string)) {
	for k := range s.m {
		iter(k)
	}
}

func NewInMemoryStorage() PluginStorage {
	return &InMemoryStorage{
		m: map[string]string{},
	}
}
