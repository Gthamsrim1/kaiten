// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/persist"
)

func commitTestFS(t *testing.T, repo string, fs *KaitenFS) {
	t.Helper()

	if err := persist.Commit(repo, fs); err != nil {
		t.Fatal(err)
	}
}

func TestRestoreEmpty(t *testing.T) {
	fs := newTestFS()
	dir := t.TempDir()

	commitTestFS(t, dir, fs)

	restored, err := Restore(dir)
	if err != nil {
		t.Fatal(err)
	}

	if restored.Root == nil {
		t.Fatal("expected root")
	}

	if len(restored.Root.Children) != 0 {
		t.Fatalf("expected empty root, got %d children", len(restored.Root.Children))
	}
}

func TestRestoreSingleFile(t *testing.T) {
	fs := newTestFS()
	dir := t.TempDir()

	_, err := fs.Root.CreateFile("hello", content.Memory([]byte("Madoka")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	commitTestFS(t, dir, fs)

	restored, err := Restore(dir)
	if err != nil {
		t.Fatal(err)
	}

	n, ok := restored.Root.Children["hello"]
	if !ok {
		t.Fatal("missing restored file")
	}

	file, ok := n.(*File)
	if !ok {
		t.Fatal("restored node is not a file")
	}

	mem := file.Content
	data, err := mem.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Madoka" {
		t.Fatalf("expected %q, got %q", "Madoka", string(data))
	}

	if file.Node.Parent != restored.Root {
		t.Fatal("incorrect parent")
	}
}

func TestRestoreNestedDirectories(t *testing.T) {
	fs := newTestFS()

	usr, err := fs.Root.CreateDirectory("usr", 0755)
	if err != nil {
		t.Fatal(err)
	}

	bin, err := usr.CreateDirectory("bin", 0755)
	if err != nil {
		t.Fatal(err)
	}

	_, err = bin.CreateFile("ls", content.Memory([]byte("binary")), 0755)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	commitTestFS(t, dir, fs)

	restored, err := Restore(dir)
	if err != nil {
		t.Fatal(err)
	}

	usrNode := restored.Root.Children["usr"].(*Directory)
	binNode := usrNode.Children["bin"].(*Directory)
	lsNode := binNode.Children["ls"].(*File)

	if usrNode.Node.Parent != restored.Root {
		t.Fatal("usr parent incorrect")
	}

	if binNode.Node.Parent != usrNode {
		t.Fatal("bin parent incorrect")
	}

	if lsNode.Node.Parent != binNode {
		t.Fatal("ls parent incorrect")
	}

	data, err := lsNode.Content.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "binary" {
		t.Fatal("file contents not restored")
	}
}

func TestRestorePreservesMetadata(t *testing.T) {
	fs := newTestFS()

	file, err := fs.Root.CreateFile("secret", content.Memory([]byte("abc")), 0600)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	commitTestFS(t, dir, fs)

	restored, err := Restore(dir)
	if err != nil {
		t.Fatal(err)
	}

	newFile := restored.Root.Children["secret"].(*File)

	if newFile.Node.ID != file.Node.ID {
		t.Fatal("id not preserved")
	}

	if newFile.Node.Mode != file.Node.Mode {
		t.Fatal("mode not preserved")
	}

	if newFile.Node.UID != file.Node.UID {
		t.Fatal("uid not preserved")
	}

	if newFile.Node.GID != file.Node.GID {
		t.Fatal("gid not preserved")
	}

	if !newFile.Node.Mtime.Equal(file.Node.Mtime) {
		t.Fatal("mtime not preserved")
	}

	if !newFile.Node.Ctime.Equal(file.Node.Ctime) {
		t.Fatal("ctime not preserved")
	}

	if !newFile.Node.Atime.Equal(file.Node.Atime) {
		t.Fatal("atime not preserved")
	}
}

func TestRestoreMissingObject(t *testing.T) {
	fs := newTestFS()

	_, err := fs.Root.CreateFile("hello", content.Memory([]byte("Madoka")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	commitTestFS(t, dir, fs)

	objectDir := filepath.Join(dir, "objects")

	entries, err := os.ReadDir(objectDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) == 0 {
		t.Fatal("expected objects")
	}

	for _, entry := range entries {
		if err := os.Remove(filepath.Join(objectDir, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRestoreLazyContentMissingObject(t *testing.T) {
	repo := t.TempDir()

	fs := newTestFS()

	root := fs.Root

	file, err := root.CreateFile("hello", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Content.Write(0, []byte("Hello Kaiten!")); err != nil {
		t.Fatal(err)
	}

	commitTestFS(t, repo, fs)

	entries, err := os.ReadDir(filepath.Join(repo, "objects"))
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if err := os.Remove(filepath.Join(repo, "objects", e.Name())); err != nil {
			t.Fatal(err)
		}
	}

	restored, err := Restore(repo)
	if err != nil {
		t.Fatalf("restore failed: %v", err)
	}

	f := restored.Root.Children["hello"].(*File)

	buf := make([]byte, f.Content.Size())

	_, err = f.Content.Read(0, buf)
	if err == nil {
		t.Fatal("expected read to fail")
	}
}
