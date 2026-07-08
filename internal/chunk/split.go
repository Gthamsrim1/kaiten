package chunk

import "crypto/sha256"

func Split(data []byte, p Params) ([]Chunk, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	bh := NewBuzhash(p.Window)

	mask := uint64(p.AvgSize - 1)

	var (
		start  int
		chunks []Chunk
	)

	for i, b := range data {
		bh.Roll(b)

		if !bh.Ready() {
			continue
		}

		size := i + 1 - start

		if size < p.MinSize {
			continue
		}

		if size >= p.MaxSize || (bh.Sum64()&mask) == 0 {
			chunks = append(chunks, makeChunk(data[start:i+1], start))
			start = i + 1
		}
	}

	if start < len(data) {
		chunks = append(chunks, makeChunk(data[start:], start))
	}

	return chunks, nil
}

func makeChunk(data []byte, offset int) Chunk {
	sum := sha256.Sum256(data)

	return Chunk{
		Offset: int64(offset),
		Hash:   sum,
		Data:   append([]byte(nil), data...),
	}
}
