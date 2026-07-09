package tree

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/persist"
)

func TestRestoreEmpty(t *testing.T) {
	fs := newTestFS()

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	if err := persist.Save(dir, snap); err != nil {
		t.Fatal(err)
	}

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

	_, err := fs.Root.CreateFile("hello", content.Memory([]byte("Madoka")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	if err := persist.Save(dir, snap); err != nil {
		t.Fatal(err)
	}

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

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	if err := persist.Save(dir, snap); err != nil {
		t.Fatal(err)
	}

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

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	if err := persist.Save(dir, snap); err != nil {
		t.Fatal(err)
	}

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

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	if err := persist.Save(dir, snap); err != nil {
		t.Fatal(err)
	}

	if len(snap.Objects) == 0 {
		t.Fatal("expected object")
	}

	if err := os.Remove(filepath.Join(dir, "objects", hex.EncodeToString(snap.Objects[0].ID[:]))); err != nil {
		t.Fatal(err)
	}

	restored, err := Restore(dir)
	if err != nil {
		t.Fatalf("restore should succeed: %v", err)
	}

	restoredFile, ok := restored.Root.Children["hello"].(*File)
	if !ok {
		t.Fatal("expected restored file")
	}

	buf := make([]byte, restoredFile.Content.Size())

	if _, err := restoredFile.Content.Read(0, buf); err == nil {
		t.Fatal("expected read to fail")
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

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if err := persist.Save(repo, snap); err != nil {
		t.Fatal(err)
	}

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