package content

import "testing"

func BenchmarkMemoryRead(b *testing.B) {
	m := Memory(make([]byte, 4096))
	buf := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Read(0, buf)
	}
}

func BenchmarkMemoryWrite(b *testing.B) {
	m := Memory(nil)
	data := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Write(0, data)
	}
}

func BenchmarkMemoryWriteAppend(b *testing.B) {
	m := Memory(nil)
	chunk := make([]byte, 64)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Write(int64(i*64), chunk)
	}
}

func BenchmarkMemorySize(b *testing.B) {
	m := Memory(make([]byte, 4096))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Size()
	}
}

// Exercises the RWMutex under contention — Read should scale roughly
// linearly with cores since it's RLock; if it doesn't, that's a signal
func BenchmarkMemoryReadParallel(b *testing.B) {
	m := Memory(make([]byte, 4096))

	b.RunParallel(func(pb *testing.PB) {
		buf := make([]byte, 4096)
		for pb.Next() {
			_, _ = m.Read(0, buf)
		}
	})
}

func BenchmarkMemoryWriteParallel(b *testing.B) {
	m := Memory(make([]byte, 4096))

	b.RunParallel(func(pb *testing.PB) {
		data := make([]byte, 64)
		for pb.Next() {
			_, _ = m.Write(0, data)
		}
	})
}