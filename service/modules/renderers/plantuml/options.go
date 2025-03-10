package plantuml

import (
	"context"
)

type Option func(*_PlantUMLRenderer)

func WithCache(getter func(key string, loader func(ctx context.Context) ([]byte, error)) ([]byte, error)) Option {
	return func(pu *_PlantUMLRenderer) {
		pu.cache = getter
	}
}

type _Cache struct {
	Light []byte `json:"light"`
	Dark  []byte `json:"dark"`
}
