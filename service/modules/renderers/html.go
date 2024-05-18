package renderers

type HTML struct {
	Renderer
}

func (me *HTML) Render(source string) (string, error) {
	return source, nil
}
