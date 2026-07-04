package tree

import (
	"context"
	"syscall"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestNewFile(t *testing.T) {
	fs := newTestFS()

	file, err := fs.Root.CreateFile("file", content.Memory([]byte("Madoka")))
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

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Madoka")))

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
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory(nil))

	n, errno := file.Write(context.Background(), nil, []byte("Madoka"), 0)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	if n != 6 {
		t.Fatalf("expected 6 bytes written, got %d", n)
	}

	mem := file.Content.(*content.MemoryContent)

	if string(mem.Bytes()) != "Madoka" {
		t.Fatalf("expected %q, got %q", "Madoka", string(mem.Bytes()))
	}
}

func TestFileOverwrite(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Homura")))

	_, errno := file.Write(context.Background(), nil, []byte("M"), 0)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	mem := file.Content.(*content.MemoryContent)

	if string(mem.Bytes()) != "Momura" {
		t.Fatalf("expected %q, got %q", "Momura", string(mem.Bytes()))
	}
}

func TestFileAppend(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Kyoko")))

	_, errno := file.Write(context.Background(), nil, []byte(" Sakura"), 5)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	mem := file.Content.(*content.MemoryContent)

	if string(mem.Bytes()) != "Kyoko Sakura" {
		t.Fatalf("expected %q, got %q", "Kyoko Sakura", string(mem.Bytes()))
	}
}

func TestFileOpen(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory(nil))

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
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory([]byte("Homura")))

	var out fuse.AttrOut

	errno := file.Getattr(context.Background(), nil, &out)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	if out.Mode != syscall.S_IFREG | 0644 {
		t.Fatalf("unexpected mode %o", out.Mode)
	}

	if out.Size != 6 {
		t.Fatalf("expected size 6, got %d", out.Size)
	}
}

func TestEmptyFile(t *testing.T) {
	fs := newTestFS()

	file, _ := fs.Root.CreateFile("file", content.Memory(nil))

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