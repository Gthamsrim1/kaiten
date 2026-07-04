package fs

import (
	"os"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestNewNode(t *testing.T) {
	fs := New()
	parent, err := fs.createDirectory("parent", fs.Root)
	if err != nil {
		t.Fatal("Couldn't create parent")
	}
	mode := uint32(fuse.S_IFREG | 0644)

	before := time.Now()
	node := newNode(fs, "hello", parent, mode)
	after := time.Now()

	if node.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	if node.Name != "hello" {
		t.Fatalf("expected name %q, got %q", "hello", node.Name)
	}

	if node.Parent != parent {
		t.Fatal("parent not set correctly")
	}

	if node.Mode != mode {
		t.Fatalf("expected mode %#o, got %#o", mode, node.Mode)
	}

	if node.UID != uint32(os.Getuid()) {
		t.Fatalf("expected UID %d, got %d", os.Getuid(), node.UID)
	}

	if node.GID != uint32(os.Getgid()) {
		t.Fatalf("expected GID %d, got %d", os.Getgid(), node.GID)
	}

	if node.Atime.Before(before) || node.Atime.After(after) {
		t.Fatal("Atime not initialized correctly")
	}

	if node.Mtime.Before(before) || node.Mtime.After(after) {
		t.Fatal("Mtime not initialized correctly")
	}

	if node.Ctime.Before(before) || node.Ctime.After(after) {
		t.Fatal("Ctime not initialized correctly")
	}
}

func TestNodeIDUnique(t *testing.T) {
	fs := New()

	n1 := newNode(fs, "one", nil, fuse.S_IFREG | 0644)
	n2 := newNode(fs, "two", nil, fuse.S_IFREG | 0644)
	n3 := newNode(fs, "three", nil, fuse.S_IFREG | 0644)

	if n1.ID == n2.ID || n2.ID == n3.ID || n1.ID == n3.ID {
		t.Fatal("node IDs must be unique")
	}

	if !(n1.ID < n2.ID && n2.ID < n3.ID) {
		t.Fatal("node IDs should increase monotonically")
	}
}