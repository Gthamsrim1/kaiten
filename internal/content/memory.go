package content

import "sync"

type MemoryContent struct {
	mu   sync.RWMutex
	data []byte
}

func Memory(data []byte) *MemoryContent {
	return &MemoryContent{
		data: append([]byte(nil), data...),
	}
}

func (m *MemoryContent) Read(offset int64, p []byte) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset >= int64(len(m.data)) {
		return 0, nil
	}

	n := copy(p, m.data[offset:])
	return n, nil
}

func (m *MemoryContent) Write(offset int64, p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	end := int(offset) + len(p)
	if end > len(m.data) {
		m.data = append(m.data, make([]byte, end-len(m.data))...)
	}

	copy(m.data[offset:], p)
	return len(p), nil
}

func (m *MemoryContent) Size() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return uint64(len(m.data))
}

// Bytes returns a copy of the underlying data. Intended for tests/debugging —
// prefer Read/Size for normal I/O.
func (m *MemoryContent) Bytes() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]byte, len(m.data))
	copy(out, m.data)
	return out
}