package post_translators

type PostTranslator interface {
	Translate(source string, base string) (string, error)
}
