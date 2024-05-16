package plantuml

type Option func(*_PlantUMLRenderer)

func WithCache(getter func(key string, loader func() (any, error)) (any, error)) Option {
	return func(pu *_PlantUMLRenderer) {
		pu.cache = getter
	}
}

type _Cache struct {
	light []byte
	dark  []byte
}
