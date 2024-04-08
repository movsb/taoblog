package renderers

type Renderer interface {
	Render(source string) (string, string, error)
}

type PathResolver interface {
	Resolve(path string) (string, error)
}
