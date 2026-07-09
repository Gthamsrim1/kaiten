// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package content

type MemoryContent struct {
	backing *Backing
}

func Memory(data []byte) *MemoryContent {
	b := &Backing{
		loaded: true,
		data:   append([]byte(nil), data...),
	}

	b.refs.Store(1)

	return &MemoryContent{
		backing: b,
	}
}

func (m *MemoryContent) Read(offset int64, p []byte) (int, error) {
	return m.backing.Read(offset, p)
}

func (m *MemoryContent) Write(offset int64, p []byte) (int, error) {
	if err := m.detach(); err != nil {
		return 0, err
	}

	return m.backing.Write(offset, p)
}

func (m *MemoryContent) Size() uint64 {
	return m.backing.Size()
}

func (m *MemoryContent) Resize(size uint64) error {
	if err := m.detach(); err != nil {
		return err
	}

	return m.backing.Resize(size)
}

func (m *MemoryContent) Bytes() ([]byte, error) {
	return m.backing.Bytes()
}

func (m *MemoryContent) Backing() *Backing {
	return m.backing
}

func (m *MemoryContent) detach() error {
	if m.backing.refs.Load() == 1 {
		return nil
	}

	data, err := m.backing.Bytes()
	if err != nil {
		return err
	}

	m.backing.Release()

	newBacking := &Backing{
		loaded: true,
		data:   data,
	}
	newBacking.refs.Store(1)

	m.backing = newBacking
	return nil
}
