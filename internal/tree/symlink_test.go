package tree_test

import (
	"syscall"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/tree"
)

func TestCreateSymlink(t *testing.T) {
	fs := tree.New()

	link, err := fs.Root.CreateSymlink("sh", "/bin/busybox")
	if err != nil {
		t.Fatal(err)
	}

	if link.Target != "/bin/busybox" {
		t.Fatalf("expected target %q, got %q", "/bin/busybox", link.Target)
	}

	if link.Node.Mode & syscall.S_IFMT != syscall.S_IFLNK {
		t.Fatalf("expected symlink mode")
	}
}

func TestCreateSymlinkAddsChild(t *testing.T) {
	fs := tree.New()

	link, err := fs.Root.CreateSymlink("sh", "/bin/busybox")
	if err != nil {
		t.Fatal(err)
	}

	children := fs.Root.ChildrenSnapshot()

	node, ok := children["sh"]
	if !ok {
		t.Fatal("symlink missing from children")
	}

	if node != link {
		t.Fatal("stored child differs from returned symlink")
	}
}

func TestCreateDuplicateSymlinkFails(t *testing.T) {
	fs := tree.New()

	_, err := fs.Root.CreateSymlink("sh", "/bin/busybox")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Root.CreateSymlink("sh", "/bin/bash"); err == nil {
		t.Fatal("expected duplicate creation to fail")
	}
}