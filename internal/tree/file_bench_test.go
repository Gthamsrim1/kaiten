package tree

import (
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func BenchmarkFileRead(b *testing.B) {
	fs := newTestFS()
	file, _ := fs.Root.CreateFile("file", content.Memory(make([]byte, 4096)), 0644)
	buf := make([]byte, 4096)
	ctx := testContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file.Read(ctx, nil, buf, 0)
	}
}

func BenchmarkFileWrite(b *testing.B) {
	fs := newTestFS()
	file, _ := fs.Root.CreateFile("file", content.Memory(nil), 0644)
	data := make([]byte, 4096)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = file.Write(testContext(), nil, data, 0)
	}
}

func BenchmarkFileGetattr(b *testing.B) {
	fs := newTestFS()
	file, _ := fs.Root.CreateFile("file", content.Memory(make([]byte, 4096)), 0644)
	ctx := testContext()
	var out fuse.AttrOut

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file.Getattr(ctx, nil, &out)
	}
}

// End-to-end small-write pattern: many small appends, similar to how a
// text editor or log writer would hit the file.
func BenchmarkFileAppendPattern(b *testing.B) {
	fs := newTestFS()
	file, _ := fs.Root.CreateFile("file", content.Memory(nil), 0644)
	chunk := []byte("line of text\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = file.Write(testContext(), nil, chunk, int64(i*len(chunk)))
	}
}