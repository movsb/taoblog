package syncer_test

import (
	"testing"

	"github.com/movsb/taoblog/modules/utils/syncer"
)

type Path string

func (p Path) Less(than Path) bool {
	return p < than
}

func (p Path) DeepEqual(to Path) bool {
	return p == to
}

// TODO 用 slices.EqualFunc 完成对比。
func TestSync(t *testing.T) {
	deletedLocal := []Path{}
	deletedRemote := []Path{}
	copiedFromLocalToRemote := []Path{}
	copiedFromRemoteToLocal := []Path{}

	s := syncer.New(
		syncer.WithDeleteLocal[[]Path](func(p Path) error {
			deletedLocal = append(deletedLocal, p)
			return nil
		}),
		syncer.WithDeleteRemote[[]Path](func(p Path) error {
			deletedRemote = append(deletedRemote, p)
			return nil
		}),
		syncer.WithCopyLocalToRemote[[]Path](func(p Path) error {
			copiedFromLocalToRemote = append(copiedFromLocalToRemote, p)
			return nil
		}),
		syncer.WithCopyRemoteToLocal[[]Path](func(p Path) error {
			copiedFromRemoteToLocal = append(copiedFromRemoteToLocal, p)
			return nil
		}),
	)

	locals := []Path{"a", "b", "c"}
	remotes := []Path{"b", "c", "d"}

	clear2 := func() {
		clear(deletedLocal)
		clear(deletedRemote)
		clear(copiedFromLocalToRemote)
		clear(copiedFromRemoteToLocal)
	}

	clear2()
	err := s.Sync(locals, remotes, syncer.LocalToRemote)
	if err != nil {
		panic(err)
	}
	t.Log(deletedLocal, deletedRemote, copiedFromLocalToRemote, copiedFromRemoteToLocal)

	clear2()
	err = s.Sync(locals, remotes, syncer.RemoteToLocal)
	if err != nil {
		panic(err)
	}
	t.Log(deletedLocal, deletedRemote, copiedFromLocalToRemote, copiedFromRemoteToLocal)
}
