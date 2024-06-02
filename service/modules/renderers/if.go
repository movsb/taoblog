package renderers

type Renderer interface {
	Render(source string) (string, error)
}
