package fs

func Memory(data []byte) *MemoryContent {
	return &MemoryContent{
		data: append([]byte(nil), data...),
	}
}

func (m *MemoryContent) Read(offset int64, p []byte) (int, error) {
	if offset >= int64(len(m.data)) {
		return 0, nil
	}

	n := copy(p, m.data[offset:])
	return n, nil
}

func (m *MemoryContent) Write(offset int64, p []byte) (int, error) {
	end := int(offset) + len(p)

	if end > len(m.data) {
		newData := make([]byte, end)
		copy(newData, m.data)
		m.data = newData
	}

	copy(m.data[offset:], p)
	return len(p), nil
}

func (m *MemoryContent) Size() uint64 {
	return uint64(len(m.data))
}
