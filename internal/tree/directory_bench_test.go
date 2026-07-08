package tree

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func BenchmarkCreateFile(b *testing.B) {
	fs := newTestFS()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// name must be unique per iteration — CreateFile errors on collision
		name := fmt.Sprintf("file-%d", i)
		_, _ = fs.Root.CreateFile(name, content.Memory(nil), 0644)
	}
}

func BenchmarkCreateDirectory(b *testing.B) {
	fs := newTestFS()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("dir-%d", i)
		_, _ = fs.Root.CreateDirectory(name, 0755)
	}
}

func BenchmarkDeleteFile(b *testing.B) {
	fs := newTestFS()
	names := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		names[i] = fmt.Sprintf("file-%d", i)
		_, _ = fs.Root.CreateFile(names[i], content.Memory(nil), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.Root.DeleteFile(names[i])
	}
}

// Lookup performance as the directory grows — run with -bench=Lookup
// and compare across sizes to see how map lookup + Mount overhead scales.
func benchmarkLookup(b *testing.B, n int) {
	fs := newTestFS()
	for i := 0; i < n; i++ {
		_, _ = fs.Root.CreateFile(fmt.Sprintf("file-%d", i), content.Memory(nil), 0644)
	}
	ctx := testContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out fuse.EntryOut
		_, _ = fs.Root.Lookup(ctx, "file-0", &out)
	}
}

func BenchmarkLookup_10(b *testing.B) {
	benchmarkLookup(b, 10)
}
func BenchmarkLookup_1000(b *testing.B) {
	benchmarkLookup(b, 1000)
}
func BenchmarkLookup_10000(b *testing.B) {
	benchmarkLookup(b, 10000)
}

// Readdir cost as a function of directory size — the DirStream allocation
// and Children map iteration are the two things worth watching here.
func benchmarkReaddir(b *testing.B, n int) {
	fs := newTestFS()
	for i := 0; i < n; i++ {
		_, _ = fs.Root.CreateFile(fmt.Sprintf("file-%d", i), content.Memory(nil), 0644)
	}
	ctx := testContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fs.Root.Readdir(ctx)
	}
}

func BenchmarkReaddir_10(b *testing.B) {
	benchmarkReaddir(b, 10)
}
func BenchmarkReaddir_1000(b *testing.B) {
	benchmarkReaddir(b, 1000)
}
func BenchmarkReaddir_10000(b *testing.B) {
	benchmarkReaddir(b, 10000)
}

// Concurrent create load — exercises the Children mutex under real
// contention, each goroutine using a unique name via atomic counter.
func BenchmarkCreateFileParallel(b *testing.B) {
	fs := newTestFS()
	var counter int64

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := atomic.AddInt64(&counter, 1)
			name := fmt.Sprintf("file-%d", id)
			_, _ = fs.Root.CreateFile(name, content.Memory(nil), 0644)
		}
	})
}

func BenchmarkRename(b *testing.B) {
	fs := newTestFS()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("file%d", i)
		if _, err := fs.Root.CreateFile(name, content.Memory(nil), 0644); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		oldName := fmt.Sprintf("file%d", i)
		newName := fmt.Sprintf("renamed%d", i)
		if err := fs.rename(fs.Root, fs.Root, oldName, newName); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenameMove ping-pongs a single file between two directories,
// exercising the cross-parent path without growing memory with b.N.
func BenchmarkRenameMove(b *testing.B) {
	fs := newTestFS()
	src, err := fs.Root.CreateDirectory("src", 0755)
	if err != nil {
		b.Fatal(err)
	}
	dst, err := fs.Root.CreateDirectory("dst", 0755)
	if err != nil {
		b.Fatal(err)
	}
	if _, err := src.CreateFile("file", content.Memory(nil), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	from, to := src, dst
	for i := 0; i < b.N; i++ {
		if err := fs.rename(from, to, "file", "file"); err != nil {
			b.Fatal(err)
		}
		from, to = to, from
	}
}