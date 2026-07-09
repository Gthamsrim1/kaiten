// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree

import (
	"syscall"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/errs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func TestCreateFile(t *testing.T) {
	fs := newTestFS()
	file, err := fs.Root.CreateFile("file", content.Memory([]byte("Hello")), 0644)
	if err != nil {
		t.Fatal(err)
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

func TestDeleteFile(t *testing.T) {
	fs := newTestFS()
	_, err := fs.Root.CreateFile("file", content.Memory([]byte("Hello")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Root.DeleteFile("file")
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteDirectory(t *testing.T) {
	fs := newTestFS()
	_, err := fs.Root.CreateDirectory("dir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Root.DeleteDirectory("dir")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateDirectory(t *testing.T) {
	fs := newTestFS()
	dir, err := fs.Root.CreateDirectory("directory", 0755)
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
	fs := newTestFS()
	parent, _ := fs.Root.CreateDirectory("parent", 0755)
	child, _ := parent.CreateDirectory("child", 0755)

	if _, ok := parent.Children[child.Node.Name]; !ok {
		t.Fatalf("Failed to allocate children")
	}
}

func TestChildrenMapInitialized(t *testing.T) {
	fs := newTestFS()

	parent, err := fs.Root.CreateDirectory("parent", 0755)
	if err != nil {
		t.Fatal(err)
	}

	if parent.Children == nil {
		t.Fatal("children map was not initialized")
	}
}

func TestReaddir(t *testing.T) {
	fs := newTestFS()

	_, err := fs.Root.CreateFile("hello", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateFile("readme", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateDirectory("docs", 0755)
	if err != nil {
		t.Fatal(err)
	}

	ctx := testContext()

	stream, errno := fs.Root.Readdir(ctx)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	expected := map[string]uint32{
		"hello":  syscall.S_IFREG | 0644,
		"readme": syscall.S_IFREG | 0644,
		"docs":   syscall.S_IFDIR | 0755,
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
	fs := newTestFS()
	_, _ = fs.Root.CreateFile("file1", content.Memory([]byte("Hello")), 0644)
	_, err := fs.Root.CreateFile("file1", content.Memory([]byte("Hello")), 0644)
	if err == nil {
		t.Fatal("Expected error: Duplicate Files")
	}
}

func TestCreateDuplicateDirectory(t *testing.T) {
	fs := newTestFS()
	_, _ = fs.Root.CreateDirectory("directory", 0755)
	_, err := fs.Root.CreateDirectory("directory", 0755)
	if err == nil {
		t.Fatal("Expected error: Duplicate Directories")
	}
}

func TestDeleteMissingFile(t *testing.T) {
	fs := newTestFS()

	err := fs.Root.DeleteFile("file")
	if err == nil {
		t.Fatal("Expected File not found, got deleted")
	}
}

func TestDeleteMissingDir(t *testing.T) {
	fs := newTestFS()

	err := fs.Root.DeleteDirectory("dir")
	if err == nil {
		t.Fatal("Expected Directory not found, got deleted")
	}
}

func TestDeleteDirectoryAsFile(t *testing.T) {
	fs := newTestFS()
	_, err := fs.Root.CreateDirectory("file", 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Root.DeleteFile("file")
	if err == nil {
		t.Fatal("Expected File not found, got deleted")
	}
}

func TestDeleteFileAsDirectory(t *testing.T) {
	fs := newTestFS()
	_, err := fs.Root.CreateFile("file", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = fs.Root.DeleteDirectory("file")
	if err == nil {
		t.Fatal("Expected Directory not found, got deleted")
	}
}

func TestRenameFileSameDirectory(t *testing.T) {
	fs := newTestFS()
	file, err := fs.Root.CreateFile("old", content.Memory([]byte("data")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := fs.rename(fs.Root, fs.Root, "old", "new"); err != nil {
		t.Fatal(err)
	}

	if _, ok := fs.Root.Children["old"]; ok {
		t.Fatal("old name still present")
	}

	got, ok := fs.Root.Children["new"]
	if !ok {
		t.Fatal("new name not present")
	}
	if got != file {
		t.Fatal("renamed node is not the same instance")
	}
	if file.Node.Name != "new" {
		t.Fatalf("expected name %q, got %q", "new", file.Node.Name)
	}
	if file.Node.Parent != fs.Root {
		t.Fatal("parent should remain unchanged for a same-directory rename")
	}
}

func TestRenameDirectorySameParent(t *testing.T) {
	fs := newTestFS()
	dir, err := fs.Root.CreateDirectory("olddir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	if err := fs.rename(fs.Root, fs.Root, "olddir", "newdir"); err != nil {
		t.Fatal(err)
	}

	if _, ok := fs.Root.Children["olddir"]; ok {
		t.Fatal("old dir name still present")
	}
	if _, ok := fs.Root.Children["newdir"]; !ok {
		t.Fatal("new dir name not present")
	}
	if dir.Node.Name != "newdir" {
		t.Fatalf("expected name %q, got %q", "newdir", dir.Node.Name)
	}
}

func TestRenameMoveToDifferentParent(t *testing.T) {
	fs := newTestFS()
	src, err := fs.Root.CreateDirectory("src", 0755)
	if err != nil {
		t.Fatal(err)
	}
	dst, err := fs.Root.CreateDirectory("dst", 0755)
	if err != nil {
		t.Fatal(err)
	}

	file, err := src.CreateFile("file", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if err := fs.rename(src, dst, "file", "file"); err != nil {
		t.Fatal(err)
	}

	if _, ok := src.Children["file"]; ok {
		t.Fatal("file still present in source directory")
	}
	moved, ok := dst.Children["file"]
	if !ok {
		t.Fatal("file not present in destination directory")
	}
	if moved != file {
		t.Fatal("moved node is not the same instance")
	}
	if file.Node.Parent != dst {
		t.Fatal("parent was not updated to the destination directory")
	}
}

func TestRenameNonexistentSource(t *testing.T) {
	fs := newTestFS()

	err := fs.rename(fs.Root, fs.Root, "missing", "new")
	if err != errs.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRenameInvalidNewName(t *testing.T) {
	fs := newTestFS()
	if _, err := fs.Root.CreateFile("file", content.Memory(nil), 0644); err != nil {
		t.Fatal(err)
	}

	err := fs.rename(fs.Root, fs.Root, "file", "..")
	if err == nil {
		t.Fatal("expected error for invalid new name")
	}
}

func TestRenameOverEmptyDirectory(t *testing.T) {
	fs := newTestFS()
	if _, err := fs.Root.CreateDirectory("src", 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Root.CreateDirectory("dst", 0755); err != nil {
		t.Fatal(err)
	}

	if err := fs.rename(fs.Root, fs.Root, "src", "dst"); err != nil {
		t.Fatal(err)
	}

	if _, ok := fs.Root.Children["src"]; ok {
		t.Fatal("source name still present")
	}
	if _, ok := fs.Root.Children["dst"]; !ok {
		t.Fatal("destination name missing after rename")
	}
}

func TestRenameOverNonEmptyDirectory(t *testing.T) {
	fs := newTestFS()
	if _, err := fs.Root.CreateDirectory("src", 0755); err != nil {
		t.Fatal(err)
	}
	dst, err := fs.Root.CreateDirectory("dst", 0755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dst.CreateFile("inner", content.Memory(nil), 0644); err != nil {
		t.Fatal(err)
	}

	err = fs.rename(fs.Root, fs.Root, "src", "dst")
	if err != errs.ErrNotEmpty {
		t.Fatalf("expected ErrNotEmpty, got %v", err)
	}
}

func TestRenameDirectoryOverFile(t *testing.T) {
	fs := newTestFS()
	if _, err := fs.Root.CreateDirectory("src", 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Root.CreateFile("dst", content.Memory(nil), 0644); err != nil {
		t.Fatal(err)
	}

	err := fs.rename(fs.Root, fs.Root, "src", "dst")
	if err != errs.ErrNotDirectory {
		t.Fatalf("expected ErrNotDirectory, got %v", err)
	}
}

func TestRenameFileOverFile(t *testing.T) {
	fs := newTestFS()
	src, err := fs.Root.CreateFile("src", content.Memory([]byte("source")), 0644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.Root.CreateFile("dst", content.Memory([]byte("dest")), 0644); err != nil {
		t.Fatal(err)
	}

	if err := fs.rename(fs.Root, fs.Root, "src", "dst"); err != nil {
		t.Fatal(err)
	}

	got, ok := fs.Root.Children["dst"]
	if !ok {
		t.Fatal("dst missing after rename")
	}
	if got != src {
		t.Fatal("dst should now point at the renamed source file")
	}
}

func TestLookupExistingFile(t *testing.T) {
	fs := newTestFS()

	_, _ = fs.Root.CreateFile("hello", content.Memory(nil), 0644)

	var out fuse.EntryOut

	inode, errno := fs.Root.Lookup(testContext(), "hello", &out)
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	if inode == nil {
		t.Fatal("expected inode")
	}

	if out.Attr.Mode != syscall.S_IFREG|0644 {
		t.Fatalf("unexpected mode %o", out.Attr.Mode)
	}
}

func TestLookupMissing(t *testing.T) {
	fs := newTestFS()

	var out fuse.EntryOut

	_, errno := fs.Root.Lookup(testContext(), "missing", &out)

	if errno != syscall.ENOENT {
		t.Fatalf("expected ENOENT, got %v", errno)
	}
}

func TestDeleteRemovesChild(t *testing.T) {
	fs := newTestFS()

	_, _ = fs.Root.CreateFile("file", content.Memory(nil), 0644)

	if err := fs.Root.DeleteFile("file"); err != nil {
		t.Fatal(err)
	}

	if _, ok := fs.Root.Children["file"]; ok {
		t.Fatal("file still present after delete")
	}
}

func TestDeleteDirectoryRemovesChild(t *testing.T) {
	fs := newTestFS()

	_, _ = fs.Root.CreateDirectory("dir", 0755)

	if err := fs.Root.DeleteDirectory("dir"); err != nil {
		t.Fatal(err)
	}

	if _, ok := fs.Root.Children["dir"]; ok {
		t.Fatal("directory still present after delete")
	}
}

func TestEmptyReaddir(t *testing.T) {
	fs := newTestFS()

	stream, errno := fs.Root.Readdir(testContext())
	if errno != 0 {
		t.Fatalf("expected errno 0, got %v", errno)
	}

	if stream.HasNext() {
		t.Fatal("expected empty directory")
	}
}
