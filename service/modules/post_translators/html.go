package post_translators

type HTMLTranslator struct {
	PostTranslator
}

func (me *HTMLTranslator) Translate(source string, base string) (string, error) {
	return source, nil
}
