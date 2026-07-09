package tree

import (
	"syscall"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestNewFile(t *testing.T) {
	fs := newTestFS()

	file, err := fs.Root.CreateFile("file", content.Memory([]byte("Madoka")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if file.Name != "file" {
		t.Fatal("incorrect file name")
	}

	if file.Node.Parent != fs.Root {
		t.Fatal("incorrect parent")
	}

	if file.Content == nil {
		t.Fatal("content not initialized")
	}
}

func TestFileRead(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Madoka")), 0644)

	buf := make([]byte, 6)

	result, errno := file.Read(testContext(), nil, buf, 0)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	data, status := result.Bytes(buf)
	if status != fuse.OK {
		t.Fatalf("unexpected status: %v", status)
	}

	if string(data) != "Madoka" {
		t.Fatalf("expected %q, got %q", "Madoka", string(data))
	}
}

func TestFileWrite(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory(nil), 0644)

	n, errno := file.Write(testContext(), nil, []byte("Madoka"), 0)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	if n != 6 {
		t.Fatalf("expected 6 bytes written, got %d", n)
	}

	mem := file.Content.(*content.MemoryContent)
	data, err := mem.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Madoka" {
		t.Fatalf("expected %q, got %q", "Madoka", string(data))
	}
}

func TestFileOverwrite(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Homura")), 0644)

	_, errno := file.Write(testContext(), nil, []byte("M"), 0)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	mem := file.Content.(*content.MemoryContent)
	data, err := mem.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Momura" {
		t.Fatalf("expected %q, got %q", "Momura", string(data))
	}
}

func TestFileAppend(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Kyoko")), 0644)

	_, errno := file.Write(testContext(), nil, []byte(" Sakura"), 5)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	mem := file.Content.(*content.MemoryContent)
	data, err := mem.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Kyoko Sakura" {
		t.Fatalf("expected %q, got %q", "Kyoko Sakura", string(data))
	}
}

func TestFileOpen(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory(nil), 0644)

	fh, flags, errno := file.Open(testContext(), 0)

	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	if fh != nil {
		t.Fatal("expected nil file handle")
	}

	if flags != fuse.FOPEN_DIRECT_IO {
		t.Fatalf("expected FOPEN_DIRECT_IO, got %v", flags)
	}
}

func TestFileGetattr(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Homura")), 0644)

	var out fuse.AttrOut

	errno := file.Getattr(testContext(), nil, &out)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	if out.Mode != syscall.S_IFREG|0644 {
		t.Fatalf("unexpected mode %o", out.Mode)
	}

	if out.Size != 6 {
		t.Fatalf("expected size 6, got %d", out.Size)
	}
}

func TestEmptyFile(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory(nil), 0644)

	buf := make([]byte, 10)

	result, errno := file.Read(testContext(), nil, buf, 0)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	data, status := result.Bytes(buf)
	if status != fuse.OK {
		t.Fatalf("unexpected status: %v", status)
	}

	if len(data) != 0 {
		t.Fatalf("expected empty read, got %d bytes", len(data))
	}
}

func TestFileReadOffset(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("abcdef")), 0644)

	buf := make([]byte, 4)

	result, errno := file.Read(testContext(), nil, buf, 2)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	data, status := result.Bytes(buf)
	if status != fuse.OK {
		t.Fatalf("unexpected status %v", status)
	}

	if string(data) != "cdef" {
		t.Fatalf("expected %q, got %q", "cdef", string(data))
	}
}

func TestFileTruncateSmaller(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Hello World")), 0644)

	var in fuse.SetAttrIn
	in.Valid = fuse.FATTR_SIZE
	in.Size = 5

	var out fuse.AttrOut

	errno := file.Setattr(testContext(), nil, &in, &out)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	mem := file.Content.(*content.MemoryContent)
	data, err := mem.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Hello" {
		t.Fatalf("expected %q, got %q", "Hello", string(data))
	}

	if out.Size != 5 {
		t.Fatalf("expected size 5, got %d", out.Size)
	}
}

func TestFileExtend(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Hello")), 0644)

	var in fuse.SetAttrIn
	in.Valid = fuse.FATTR_SIZE
	in.Size = 10

	var out fuse.AttrOut

	errno := file.Setattr(testContext(), nil, &in, &out)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	mem := file.Content.(*content.MemoryContent)
	data, err := mem.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 10 {
		t.Fatalf("expected len 10, got %d", len(data))
	}

	if string(data[:5]) != "Hello" {
		t.Fatal("existing contents corrupted")
	}

	if out.Size != 10 {
		t.Fatalf("expected size 10, got %d", out.Size)
	}
}

func TestFileWriteMiddle(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("abcdef")), 0644)

	_, errno := file.Write(testContext(), nil, []byte("XYZ"), 3)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	mem := file.Content.(*content.MemoryContent)
	data, err := mem.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "abcXYZ" {
		t.Fatalf("expected %q, got %q", "abcXYZ", string(data))
	}
}

func TestReadPastEOF(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("abc")), 0644)

	buf := make([]byte, 10)

	result, errno := file.Read(testContext(), nil, buf, 100)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	data, status := result.Bytes(buf)
	if status != fuse.OK {
		t.Fatalf("unexpected status %v", status)
	}

	if len(data) != 0 {
		t.Fatalf("expected empty read, got %d bytes", len(data))
	}
}
