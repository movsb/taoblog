package renderers

type Renderer interface {
	Render(source string) (string, error)
}

type HTML struct {
	Renderer
}

func (me *HTML) Render(source string) (string, error) {
	return source, nil
}
