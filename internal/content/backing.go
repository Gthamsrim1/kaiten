// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package content

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Gthamsrim1/kaiten/internal/store"
)

type ObjectLoader interface {
	Load(hash [32]byte) ([]byte, error)
}

type Backing struct {
	mu sync.RWMutex

	loaded bool
	data   []byte

	chunks []store.ChunkRef
	loader ObjectLoader

	refs atomic.Int32
}

func (b *Backing) ensureLoaded() error {
	b.mu.RLock()
	if b.loaded {
		b.mu.RUnlock()
		return nil
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.loaded {
		return nil
	}

	if len(b.chunks) == 0 {
		b.data = []byte{}
		b.loaded = true
		return nil
	}

	var total uint64
	for _, c := range b.chunks {
		total += uint64(c.Length)
	}

	b.data = make([]byte, 0, int(total))

	for _, c := range b.chunks {
		object, err := b.loader.Load(c.Hash)
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

		b.data = append(b.data, object...)
	}

	b.loaded = true
	return nil
}

func (b *Backing) Read(offset int64, p []byte) (int, error) {
	if err := b.ensureLoaded(); err != nil {
		return 0, err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if offset >= int64(len(b.data)) {
		return 0, nil
	}

	n := copy(p, b.data[offset:])
	return n, nil
}

func (b *Backing) Write(offset int64, p []byte) (int, error) {
	if err := b.ensureLoaded(); err != nil {
		return 0, err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	end := int(offset) + len(p)
	if end > len(b.data) {
		b.data = append(b.data, make([]byte, end-len(b.data))...)
	}

	copy(b.data[offset:], p)
	return len(p), nil
}

func (b *Backing) Size() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.loaded {
		return uint64(len(b.data))
	}

	var size uint64
	for _, c := range b.chunks {
		size += uint64(c.Length)
	}

	return size
}

func (b *Backing) Resize(size uint64) error {
	if err := b.ensureLoaded(); err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	current := uint64(len(b.data))

	switch {
	case size < current:
		b.data = b.data[:size]

	case size > current:
		newData := make([]byte, size)
		copy(newData, b.data)
		b.data = newData
	}

	return nil
}

func (b *Backing) Bytes() ([]byte, error) {
	if err := b.ensureLoaded(); err != nil {
		return nil, err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]byte, len(b.data))
	copy(out, b.data)
	return out, nil
}

func (b *Backing) Acquire() {
	b.refs.Add(1)
}

func (b *Backing) Release() {
	if b.refs.Add(-1) < 0 {
		panic("negative backing refs")
	}
}

func (b *Backing) Refs() int32 {
	return b.refs.Load()
}

func FromBacking(b *Backing) *MemoryContent {
	return &MemoryContent{
		backing: b,
	}
}
