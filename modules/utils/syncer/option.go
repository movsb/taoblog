package syncer

type Option[S ~[]E, E Element[E]] interface {
	Apply(*Syncer[S, E])
}

type Op[S ~[]E, E Element[E]] struct {
	fn    func(e E) error
	index _Op
}

func (op Op[S, E]) Apply(s *Syncer[S, E]) {
	s.ops[op.index] = op.fn
}

func WithDeleteLocal[S ~[]E, E Element[E]](fn func(e E) error) Option[S, E] {
	return Op[S, E]{fn, deleteLocal}
}

func WithDeleteRemote[S ~[]E, E Element[E]](fn func(e E) error) Option[S, E] {
	return Op[S, E]{fn, deleteRemote}
}

func WithCopyLocalToRemote[S ~[]E, E Element[E]](fn func(e E) error) Option[S, E] {
	return Op[S, E]{fn, copyLocalToRemote}
}

func WithCopyRemoteToLocal[S ~[]E, E Element[E]](fn func(e E) error) Option[S, E] {
	return Op[S, E]{fn, copyRemoteToLocal}
}
