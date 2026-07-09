package content

import (
	"github.com/Gthamsrim1/kaiten/internal/store"
)

type LazyContent struct {
	backing *Backing
}

func Lazy(chunks []store.ChunkRef, loader ObjectLoader) *LazyContent {
	b := &Backing{
		chunks: chunks,
		loader: loader,
	}

	b.refs.Store(1)

	return &LazyContent{
		backing: b,
	}
}

func (l *LazyContent) Read(offset int64, p []byte) (int, error) {
	return l.backing.Read(offset, p)
}

func (l *LazyContent) Write(offset int64, p []byte) (int, error) {
	if err := l.detach(); err != nil {
		return 0, err
	}

	return l.backing.Write(offset, p)
}

func (l *LazyContent) Size() uint64 {
	return l.backing.Size()
}

func (l *LazyContent) Resize(size uint64) error {
	if err := l.detach(); err != nil {
		return err
	}

	return l.backing.Resize(size)
}

func (l *LazyContent) Bytes() ([]byte, error) {
	return l.backing.Bytes()
}

func (l *LazyContent) Backing() *Backing {
	return l.backing
}

func (l *LazyContent) detach() error {
	if l.backing.refs.Load() == 1 {
		return nil
	}

	data, err := l.backing.Bytes()
	if err != nil {
		return err
	}

	l.backing.Release()

	newBacking := &Backing{
		loaded: true,
		data:   data,
	}
	newBacking.refs.Store(1)

	l.backing = newBacking
	return nil
}
