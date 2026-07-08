package chunk

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func benchmarkSplit(b *testing.B, data []byte) {
	params := testParams()

	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := Split(data, params); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSplit64KB(b *testing.B) {
	data := make([]byte, 64*1024)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}

	benchmarkSplit(b, data)
}

func BenchmarkSplit1MB(b *testing.B) {
	data := make([]byte, 1*1024*1024)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}

	benchmarkSplit(b, data)
}

func BenchmarkSplit10MB(b *testing.B) {
	data := make([]byte, 10*1024*1024)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}

	benchmarkSplit(b, data)
}

func BenchmarkSplit100MB(b *testing.B) {
	data := make([]byte, 100*1024*1024)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}

	benchmarkSplit(b, data)
}

func BenchmarkSplit1MBRandom(b *testing.B) {
	data := make([]byte, 1*1024*1024)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}

	benchmarkSplit(b, data)
}

func BenchmarkSplit1MBRepeated(b *testing.B) {
	data := bytes.Repeat([]byte("A"), 1*1024*1024)

	benchmarkSplit(b, data)
}

func BenchmarkSplit1MBZeroes(b *testing.B) {
	data := make([]byte, 1*1024*1024)

	benchmarkSplit(b, data)
}

func BenchmarkSplit1MBAlternating(b *testing.B) {
	data := make([]byte, 1*1024*1024)

	for i := range data {
		if i%2 == 0 {
			data[i] = 0xAA
		} else {
			data[i] = 0x55
		}
	}

	benchmarkSplit(b, data)
}

func BenchmarkSplit1MBSequential(b *testing.B) {
	data := make([]byte, 1*1024*1024)

	for i := range data {
		data[i] = byte(i)
	}

	benchmarkSplit(b, data)
}
