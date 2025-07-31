package blocknote

type Option func(b *Blocknote)

func WithTitle(title *string) Option {
	return func(b *Blocknote) {
		b.title = title
	}
}

func DoNotRenderTitle() Option {
	return func(b *Blocknote) {
		b.doNotRenderTitle = true
	}
}
