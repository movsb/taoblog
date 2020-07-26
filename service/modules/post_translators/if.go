package post_translators

type PostTranslator interface {
	Translate(cb *Callback, source string, base string) (string, error)
}

// Callback ...
type Callback struct {
	SetTitle func(title string)
}
