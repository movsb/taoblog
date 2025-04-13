package plantuml

import (
	"time"

	"github.com/movsb/taoblog/service/modules/cache"
)

type Option func(*_PlantUMLRenderer)

type CacheKey struct {
	compressed string
}

func (c CacheKey) CacheKey() []byte {
	return []byte(c.compressed)
}

func WithFileCache(c cache.FileCache) Option {
	return func(pu *_PlantUMLRenderer) {
		pu.cache = func(key CacheKey, loader func() ([]byte, error)) ([]byte, error) {
			return c.GetOrLoad(key, func() ([]byte, time.Duration, error) {
				val, err := loader()
				return val, cache.OneMonth, err
			})
		}
	}
}

type _Cache struct {
	Light []byte `json:"light"`
	Dark  []byte `json:"dark"`
}
