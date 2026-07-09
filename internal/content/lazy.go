package content

import (
	"fmt"
	"sync"

	"github.com/Gthamsrim1/kaiten/internal/store"
)

type ObjectLoader interface {
    Load(hash [32]byte) ([]byte, error)
}

type LazyContent struct {
    mu sync.RWMutex

    loaded bool
    data []byte

    chunks []store.ChunkRef
    loader ObjectLoader
}

func Lazy(chunks []store.ChunkRef, loader ObjectLoader) *LazyContent {
	return &LazyContent{
		chunks: chunks,
		loader: loader,
	}
}

func (l *LazyContent) ensureLoaded() error {
	l.mu.RLock()
	if l.loaded {
		l.mu.RUnlock()
		return nil
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.loaded {
		return nil
	}

	if len(l.chunks) == 0 {
		l.data = []byte{}
		l.loaded = true
		return nil
	}

	var total uint64
	for _, c := range l.chunks {
		total += uint64(c.Length)
	}

	l.data = make([]byte, 0, int(total))

	for _, c := range l.chunks {
		object, err := l.loader.Load(c.Hash)
		if err != nil {
			return err
		}

		if uint32(len(object)) != c.Length {
			return fmt.Errorf(
				"object %x: expected %d bytes, got %d",
				c.Hash,
				c.Length,
				len(object),
			)
		}

		l.data = append(l.data, object...)
	}

	l.loaded = true
	return nil
}

func (l *LazyContent) Read(offset int64, p []byte) (int, error) {
	if err := l.ensureLoaded(); err != nil {
		return 0, err
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	if offset >= int64(len(l.data)) {
		return 0, nil
	}

	n := copy(p, l.data[offset:])
	return n, nil
}

func (l *LazyContent) Write(offset int64, p []byte) (int, error) {
	if err := l.ensureLoaded(); err != nil {
		return 0, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	end := int(offset) + len(p)
	if end > len(l.data) {
		l.data = append(l.data, make([]byte, end-len(l.data))...)
	}

	copy(l.data[offset:], p)
	return len(p), nil
}

func (l *LazyContent) Size() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.loaded {
		return uint64(len(l.data))
	}

	var size uint64
	for _, c := range l.chunks {
		size += uint64(c.Length)
	}

	return size
}

func (l *LazyContent) Resize(size uint64) error {
	if err := l.ensureLoaded(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	current := uint64(len(l.data))

	switch {
	case size < current:
		l.data = l.data[:size]

	case size > current:
		newData := make([]byte, size)
		copy(newData, l.data)
		l.data = newData
	}

	return nil
}

func (l *LazyContent) Bytes() ([]byte, error) {
	if err := l.ensureLoaded(); err != nil {
		return nil, err
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	out := make([]byte, len(l.data))
	copy(out, l.data)
	return out, nil
}