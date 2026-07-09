// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tree

import (
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
)

func TestCloneFile(t *testing.T) {
	fs := newTestFS()

	file, err := fs.Root.CreateFile("hello", content.Memory([]byte("Madoka")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	clonedNode, err := fs.Clone(file, fs.Root, "hello-copy")
	if err != nil {
		t.Fatal(err)
	}

	clone, ok := clonedNode.(*File)
	if !ok {
		t.Fatal("expected cloned node to be a file")
	}

	if clone == file {
		t.Fatal("clone should be a different file")
	}

	if clone.Node.ID == file.Node.ID {
		t.Fatal("clone should have a different inode id")
	}

	if clone.Node.Name != "hello-copy" {
		t.Fatalf("expected name hello-copy, got %q", clone.Node.Name)
	}

	if clone.Node.Parent != fs.Root {
		t.Fatal("clone has wrong parent")
	}

	data, err := clone.Content.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Madoka" {
		t.Fatalf("expected Madoka, got %q", string(data))
	}
}

func TestCloneDirectory(t *testing.T) {
	fs := newTestFS()

	dir, err := fs.Root.CreateDirectory("dir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dir.CreateFile("a.txt", content.Memory([]byte("A")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dir.CreateFile("b.txt", content.Memory([]byte("B")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	clonedNode, err := fs.Clone(dir, fs.Root, "dir-copy")
	if err != nil {
		t.Fatal(err)
	}

	clone := clonedNode.(*Directory)

	if clone.Node.ID == dir.ID {
		t.Fatal("directory id should differ")
	}

	if len(clone.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(clone.Children))
	}

	if _, ok := clone.Children["a.txt"]; !ok {
		t.Fatal("missing a.txt")
	}

	if _, ok := clone.Children["b.txt"]; !ok {
		t.Fatal("missing b.txt")
	}
}

func TestCloneExistingName(t *testing.T) {
	fs := newTestFS()

	file, err := fs.Root.CreateFile("hello", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateFile("copy", content.Memory(nil), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := fs.Clone(file, fs.Root, "copy"); err == nil {
		t.Fatal("expected clone to fail")
	}
}

func TestCloneSharesBacking(t *testing.T) {
	fs := newTestFS()

	file, err := fs.Root.CreateFile("hello", content.Memory([]byte("Madoka")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	clonedNode, err := fs.Clone(file, fs.Root, "copy")
	if err != nil {
		t.Fatal(err)
	}

	clone := clonedNode.(*File)

	if file.Content.Backing() != clone.Content.Backing() {
		t.Fatal("expected clone to share backing")
	}

	if file.Content.Backing().Refs() != 2 {
		t.Fatalf("expected 2 refs, got %d", file.Content.Backing().Refs())
	}
}

func TestCloneDetachOnWrite(t *testing.T) {
	fs := newTestFS()

	file, err := fs.Root.CreateFile("hello", content.Memory([]byte("Madoka")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	clonedNode, err := fs.Clone(file, fs.Root, "copy")
	if err != nil {
		t.Fatal(err)
	}

	clone := clonedNode.(*File)

	_, err = clone.Content.Write(0, []byte("Homura"))
	if err != nil {
		t.Fatal(err)
	}

	if file.Content.Backing() == clone.Content.Backing() {
		t.Fatal("backing should detach after write")
	}

	data, err := file.Content.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Madoka" {
		t.Fatalf("original modified: %q", string(data))
	}

	data, err = clone.Content.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	if string(data[:6]) != "Homura" {
		t.Fatalf("clone not modified: %q", string(data))
	}
}
