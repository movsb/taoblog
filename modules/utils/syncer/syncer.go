package syncer

import (
	"slices"
)

// 文件操作。
type _Op int

const (
	deleteLocal _Op = iota + 1
	deleteRemote
	copyLocalToRemote
	copyRemoteToLocal
)

// 同步方向：本地→远程，远程→本地。
type Dir int

const (
	LocalToRemote Dir = 1
	RemoteToLocal Dir = 2
)

type Element[E any] interface {
	Compare(to E) int    // 比较大小
	DeepEqual(to E) bool // 修改时间、权限也相同
}

// 对象同步。
// 用于文件本地、远程列表同步。
type Syncer[S ~[]E, E Element[E]] struct {
	ops map[_Op]func(e E) error
}

func New[S ~[]E, E Element[E]](options ...Option[S, E]) *Syncer[S, E] {
	s := &Syncer[S, E]{
		ops: map[_Op]func(e E) error{},
	}
	for _, opt := range options {
		opt.Apply(s)
	}
	return s
}

func (s *Syncer[S, E]) Sync(locals S, remotes S, dir Dir) error {
	l, r := s.sortAndUniq(locals, remotes)
	return s.sync(l, r, dir)
}

func (s *Syncer[S, E]) sortAndUniq(locals S, remotes S) (S, S) {
	slices.SortFunc(locals, func(a, b E) int {
		return a.Compare(b)
	})
	slices.SortFunc(remotes, func(a, b E) int {
		return a.Compare(b)
	})
	locals = slices.CompactFunc(locals, func(a, b E) bool { return a.Compare(b) == 0 })
	remotes = slices.CompactFunc(remotes, func(a, b E) bool { return a.Compare(b) == 0 })
	return locals, remotes
}

// 会自动排序、去重。
func (s *Syncer[S, E]) sync(locals S, remotes S, dir Dir) (err error) {
	i, j := len(locals)-1, len(remotes)-1
	forward := dir == LocalToRemote

	defer func() {
		if value := recover(); value != nil {
			err = value.(error)
		}
	}()

	exec := func(cond bool, op _Op, e E) {
		if !cond {
			return
		}
		fn, ok := s.ops[op]
		if !ok {
			return
		}
		if err := fn(e); err != nil {
			panic(err)
		}
	}

	extraLocal := func(e E) {
		exec(forward, copyLocalToRemote, e)
		exec(!forward, deleteLocal, e)
	}
	extraRemote := func(e E) {
		exec(forward, deleteRemote, e)
		exec(!forward, copyRemoteToLocal, e)
	}
	twoWay := func(li, rj E) {
		exec(forward, copyLocalToRemote, li)
		exec(!forward, copyRemoteToLocal, rj)
	}

	for {
		// ("没有更多需要比较的文件。")
		if i == -1 && j == -1 {
			break
		}

		// 说明远程有多的文件。
		if i == -1 {
			extraRemote(remotes[j])
			j--
			continue
		}

		// 说明本地有多的文件。
		if j == -1 {
			extraLocal(locals[i])
			i--
			continue
		}

		li, rj := locals[i], remotes[j]
		switch {
		case li.Compare(rj) == -1: // 说明远程有多的文件。
			extraRemote(rj)
			j--
			continue
		case rj.Compare(li) == -1: // 说明本地有多的文件。
			extraLocal(li)
			i--
			continue
		default: // 同名文件。
			if !li.DeepEqual(rj) {
				twoWay(li, rj)
			}
			i--
			j--
		}
	}

	return nil
}
