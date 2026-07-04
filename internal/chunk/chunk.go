package chunk

import (
	"crypto/sha256"
	"io"
)

const (
	MinSize = 2 * 1024
	MaxSize = 64 * 1024
	avgSize = 16 * 1024
	windowSize = 64
	boundaryMask = uint64(avgSize - 1)
)

// Chunk describes one content-defined chunk of a larger stream.
type Chunk struct {
	Hash   [32]byte
	Offset int64
	Length int64
}

var buzhashTable = func() [256]uint64 {
	var t [256]uint64
	seed := uint64(0x9E3779B97F4A7C15)
	next := func() uint64 {
		seed += 0x9E3779B97F4A7C15
		z := seed
		z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
		z = (z ^ (z >> 27)) * 0x94D049BB133111EB
		return z ^ (z >> 31)
	}
	for i := range t {
		t[i] = next()
	}
	return t
}()

func rotl(x uint64, n uint) uint64 {
	n %= 64
	if n == 0 {
		return x
	}
	return (x << n) | (x >> (64 - n))
}

// Split reads all of r and returns the content-defined chunks that cover it
func Split(r io.Reader) ([]Chunk, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return SplitBytes(buf), nil
}

// SplitBytes is the same as Split but operates on an in-memory buffer
// avoiding an io.Reader round trip when the caller already has the data
func SplitBytes(data []byte) []Chunk {
	if len(data) == 0 {
		return nil
	}

	var chunks []Chunk
	start := 0
	var h uint64
	var window [windowSize]byte
	var pos int

	for i := 0; i < len(data); i++ {
		b := data[i]
		outgoing := window[pos]
		window[pos] = b
		pos++
		if pos == windowSize {
			pos = 0
		}

		h = rotl(h, 1) ^ buzhashTable[b] ^ rotl(buzhashTable[outgoing], windowSize)

		length := i - start + 1
		atBoundary := length >= MinSize && (h&boundaryMask) == 0
		atMax := length >= MaxSize
		isLast := i == len(data)-1

		if atBoundary || atMax || isLast {
			chunkData := data[start : i+1]
			chunks = append(chunks, Chunk{
				Hash:   sha256.Sum256(chunkData),
				Offset: int64(start),
				Length: int64(len(chunkData)),
			})
			start = i + 1
		}
	}

	return chunks
}