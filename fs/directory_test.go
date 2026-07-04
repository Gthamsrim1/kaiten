package fs

import (
	"context"
	"syscall"
	"testing"
)

func TestCreateFile(t *testing.T) {
	fs := New()
	file, err := fs.Root.CreateFile("file", Memory([]byte("Hello")))
	if err != nil {
		t.Fatal("Couldn't create file")
	}

	if file.Name != "file" {
		t.Fatal("incorrect file name")
	}

	if file.Node.Parent != fs.Root {
		t.Fatal("incorrect parent")
	}

	if _, ok := fs.Root.Children["file"]; !ok {
		t.Fatal("file not added to children")
	}
}

func TestCreateDirectory(t *testing.T) {
	fs := New()
	dir, err := fs.Root.CreateDirectory("directory")
	if err != nil {
		t.Fatal("Couldn't create Directory")
	}

	if dir.Name != "directory" {
		t.Fatal("incorrect file name")
	}

	if dir.Node.Parent != fs.Root {
		t.Fatal("incorrect parent")
	}

	if _, ok := fs.Root.Children["directory"]; !ok {
		t.Fatal("file not added to children")
	}
}

func TestNewDirectory(t *testing.T) {
	fs := New()
	parent, _ := fs.Root.CreateDirectory("parent")
	child, _ := parent.CreateDirectory("child")
	
	if _, ok := parent.Children[child.Node.Name]; !ok {
		t.Fatalf("Failed to allocate children")
	}
}

func TestChildrenMapInitialized(t *testing.T) {
	fs := New()

	parent, err := fs.Root.CreateDirectory("parent")
	if err != nil {
		t.Fatal(err)
	}

	if parent.Children == nil {
		t.Fatal("children map was not initialized")
	}
}

func TestReaddir(t *testing.T) {
	fs := New()

	_, err := fs.Root.CreateFile("hello", Memory(nil))
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateFile("readme", Memory(nil))
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateDirectory("docs")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	stream, errno := fs.Root.Readdir(ctx)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	expected := map[string]uint32{
		"hello":  syscall.S_IFREG,
		"readme": syscall.S_IFREG,
		"docs":   syscall.S_IFDIR,
	}

	count := 0

	for stream.HasNext() {
		entry, errno := stream.Next()
		if errno != 0 {
			t.Fatalf("unexpected errno: %v", errno)
		}

		mode, ok := expected[entry.Name]
		if !ok {
			t.Fatalf("unexpected entry %q", entry.Name)
		}

		if entry.Mode != mode {
			t.Fatalf("expected mode %v for %q, got %v", mode, entry.Name, entry.Mode)
		}

		delete(expected, entry.Name)
		count++
	}

	if count != 3 {
		t.Fatalf("expected 3 entries, got %d", count)
	}

	if len(expected) != 0 {
		t.Fatalf("missing entries: %v", expected)
	}
}

func TestCreateDuplicateFile(t *testing.T) {
	fs := New()
	_, _ = fs.Root.CreateFile("file1", Memory([]byte("Hello")))
	_, err := fs.Root.CreateFile("file1", Memory([]byte("Hello")))
	if err == nil {
		t.Fatal("Expected error: Duplicate Files")
	}
}

func TestCreateDuplicateDirectory(t *testing.T) {
	fs := New()
	_, _= fs.Root.CreateDirectory("directory")
	_, err := fs.Root.CreateDirectory("directory")
	if err == nil {
		t.Fatal("Expected error: Duplicate Directories")
	}
}