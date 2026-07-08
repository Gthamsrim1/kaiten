package persist

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGCRemovesUnreferencedObject(t *testing.T) {
	dir := t.TempDir()

	hash1 := "object1"
	hash2 := "object2"

	fs := &Filesystem{
		Nodes: []Node{
			{
				ID:       1,
				Name:     "file",
				Type:     TypeFile,
				ObjectID: &hash1,
			},
		},
		Objects: []Object{
			{
				ID:   hash1,
				Data: []byte("Madoka"),
			},
			{
				ID:   hash2,
				Data: []byte("Homura"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hash1)); err != nil {
		t.Fatal("referenced object was deleted")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hash2)); !os.IsNotExist(err) {
		t.Fatal("unreferenced object still exists")
	}
}

func TestGCKeepsReferencedObject(t *testing.T) {
	dir := t.TempDir()

	hash := "object"

	fs := &Filesystem{
		Nodes: []Node{
			{
				ID:       1,
				Name:     "file",
				Type:     TypeFile,
				ObjectID: &hash,
			},
		},
		Objects: []Object{
			{
				ID:   hash,
				Data: []byte("Madoka"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hash)); err != nil {
		t.Fatal("referenced object was deleted")
	}
}

func TestGCEmptyRepository(t *testing.T) {
	dir := t.TempDir()

	fs := &Filesystem{}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, "objects"))
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Fatalf("expected 0 objects, got %d", len(entries))
	}
}

func TestGCMultipleReferences(t *testing.T) {
	dir := t.TempDir()

	hash := "shared"

	fs := &Filesystem{
		Nodes: []Node{
			{
				ID:       1,
				Name:     "a",
				Type:     TypeFile,
				ObjectID: &hash,
			},
			{
				ID:       2,
				Name:     "b",
				Type:     TypeFile,
				ObjectID: &hash,
			},
		},
		Objects: []Object{
			{
				ID:   hash,
				Data: []byte("Shared"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hash)); err != nil {
		t.Fatal("shared object was deleted")
	}
}

func TestGCDeletesManyObjects(t *testing.T) {
	dir := t.TempDir()

	hashA := "A"
	hashB := "B"
	hashC := "C"
	hashD := "D"

	fs := &Filesystem{
		Nodes: []Node{
			{
				ID:       1,
				Name:     "file1",
				Type:     TypeFile,
				ObjectID: &hashA,
			},
			{
				ID:       2,
				Name:     "file2",
				Type:     TypeFile,
				ObjectID: &hashC,
			},
		},
		Objects: []Object{
			{ID: hashA, Data: []byte("A")},
			{ID: hashB, Data: []byte("B")},
			{ID: hashC, Data: []byte("C")},
			{ID: hashD, Data: []byte("D")},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hashA)); err != nil {
		t.Fatal("A removed")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hashC)); err != nil {
		t.Fatal("C removed")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hashB)); !os.IsNotExist(err) {
		t.Fatal("B still exists")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hashD)); !os.IsNotExist(err) {
		t.Fatal("D still exists")
	}
}

func TestGCIgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()

	hash := "abc"

	fs := &Filesystem{
		Nodes: []Node{
			{
				ID:       1,
				Name:     "file",
				Type:     TypeFile,
				ObjectID: &hash,
			},
		},
		Objects: []Object{
			{
				ID:   hash,
				Data: []byte("Homu Homu"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	nested := filepath.Join(dir, "objects", "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(nested)
	if err != nil {
		t.Fatal(err)
	}

	if !info.IsDir() {
		t.Fatal("nested should still be a directory")
	}
}