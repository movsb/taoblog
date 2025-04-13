package plantuml

import (
	"github.com/movsb/taoblog/service/modules/cache"
)

type Option func(*_PlantUMLRenderer)

type CacheKey struct {
	Compressed string
}

func WithFileCache(c cache.Getter) Option {
	return func(pu *_PlantUMLRenderer) {
		pu.cache = c
	}
}

type _CacheValue struct {
	Light []byte
	Dark  []byte
}
