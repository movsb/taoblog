package post_translators

type PostTranslator interface {
	Translate(source string) (string, string, error)
}

type PathResolver interface {
	Resolve(path string) (string, error)
}
