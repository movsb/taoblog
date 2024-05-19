package plantuml

import "io"

type Option func(*_PlantUMLRenderer)

func WithCache(getter func(key string, loader func() (io.ReadCloser, error)) (io.ReadCloser, error)) Option {
	return func(pu *_PlantUMLRenderer) {
		pu.cache = getter
	}
}

type _Cache struct {
	Light []byte `json:"light"`
	Dark  []byte `json:"dark"`
}
