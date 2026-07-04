package fs

import (
	"context"
	"syscall"
	"testing"

	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestNewFile(t *testing.T) {
	fs := New()

	file, err := fs.Root.CreateFile("file", Memory([]byte("Madoka")))
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
	fs := New()

	file, _ := fs.Root.CreateFile("file", Memory([]byte("Madoka")))

	buf := make([]byte, 6)

	result, errno := file.Read(context.Background(), nil, buf, 0)
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
	fs := New()

	file, _ := fs.Root.CreateFile("file", Memory(nil))

	n, err := file.Write([]byte("Madoka"), 0)
	if err != nil {
		t.Fatal(err)
	}

	if n != 6 {
		t.Fatalf("expected 6 bytes written, got %d", n)
	}

	mem := file.Content.(*MemoryContent)

	if string(mem.data) != "Madoka" {
		t.Fatalf("expected %q, got %q", "Madoka", string(mem.data))
	}
}

func TestFileOverwrite(t *testing.T) {
	fs := New()

	file, _ := fs.Root.CreateFile("file", Memory([]byte("Homura")))

	_, err := file.Write([]byte("M"), 0)
	if err != nil {
		t.Fatal(err)
	}

	mem := file.Content.(*MemoryContent)

	if string(mem.data) != "Momura" {
		t.Fatalf("expected %q, got %q", "Momura", string(mem.data))
	}
}

func TestFileAppend(t *testing.T) {
	fs := New()

	file, _ := fs.Root.CreateFile("file", Memory([]byte("Kyoko")))

	_, err := file.Write([]byte(" Sakura"), 5)
	if err != nil {
		t.Fatal(err)
	}

	mem := file.Content.(*MemoryContent)

	if string(mem.data) != "Kyoko Sakura" {
		t.Fatalf("expected %q, got %q", "Kyoko Sakura", string(mem.data))
	}
}

func TestFileOpen(t *testing.T) {
	fs := New()

	file, _ := fs.Root.CreateFile("file", Memory(nil))

	fh, flags, errno := file.Open(context.Background(), 0)

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
	fs := New()

	file, _ := fs.Root.CreateFile("file", Memory([]byte("Homura")))

	var out fuse.AttrOut

	errno := file.Getattr(context.Background(), nil, &out)
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
	fs := New()

	file, _ := fs.Root.CreateFile("file", Memory(nil))

	buf := make([]byte, 10)

	result, errno := file.Read(context.Background(), nil, buf, 0)
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