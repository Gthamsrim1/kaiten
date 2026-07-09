package persist

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/store"
)

func testHash(s string) [32]byte {
	return sha256.Sum256([]byte(s))
}

func TestGCRemovesUnreferencedObject(t *testing.T) {
	dir := t.TempDir()

	hash1 := testHash("object1")
	hash2 := testHash("object2")

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID: id,
		Nodes: []Node{
			{
				ID:   1,
				Name: "file",
				Type: TypeFile,
				Chunks: []store.ChunkRef{
					{
						Hash:   hash1,
						Length: 6,
					},
				},
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

	if err := Save(dir, ss); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hash1[:]))); err != nil {
		t.Fatal("referenced object was deleted")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hash2[:]))); !os.IsNotExist(err) {
		t.Fatal("unreferenced object still exists")
	}
}

func TestGCKeepsReferencedObject(t *testing.T) {
	dir := t.TempDir()

	hash := testHash("object")

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID: id,
		Nodes: []Node{
			{
				ID:   1,
				Name: "file",
				Type: TypeFile,
				Chunks: []store.ChunkRef{
					{
						Hash:   hash,
						Length: uint32(len(hash)),
					},
				},
			},
		},
		Objects: []Object{
			{
				ID:   hash,
				Data: []byte("Madoka"),
			},
		},
	}

	if err := Save(dir, ss); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hash[:]))); err != nil {
		t.Fatal("referenced object was deleted")
	}
}

func TestGCEmptyRepository(t *testing.T) {
	dir := t.TempDir()

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID: id,
	}

	if err := Save(dir, ss); err != nil {
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

	hash := testHash("shared")

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID: id,
		Nodes: []Node{
			{
				ID:   1,
				Name: "a",
				Type: TypeFile,
				Chunks: []store.ChunkRef{
					{
						Hash:   hash,
						Length: uint32(len(hash)),
					},
				},
			},
			{
				ID:   2,
				Name: "b",
				Type: TypeFile,
				Chunks: []store.ChunkRef{
					{
						Hash:   hash,
						Length: uint32(len(hash)),
					},
				},
			},
		},
		Objects: []Object{
			{
				ID:   hash,
				Data: []byte("Shared"),
			},
		},
	}

	if err := Save(dir, ss); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hash[:]))); err != nil {
		t.Fatal("shared object was deleted")
	}
}

func TestGCDeletesManyObjects(t *testing.T) {
	dir := t.TempDir()

	hashA := testHash("A")
	hashB := testHash("B")
	hashC := testHash("C")
	hashD := testHash("D")

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID: id,
		Nodes: []Node{
			{
				ID:   1,
				Name: "file1",
				Type: TypeFile,
				Chunks: []store.ChunkRef{
					{
						Hash:   hashA,
						Length: uint32(len(hashA)),
					},
				},
			},
			{
				ID:   2,
				Name: "file2",
				Type: TypeFile,
				Chunks: []store.ChunkRef{
					{
						Hash:   hashC,
						Length: uint32(len(hashC)),
					},
				},
			},
		},
		Objects: []Object{
			{ID: hashA, Data: []byte("A")},
			{ID: hashB, Data: []byte("B")},
			{ID: hashC, Data: []byte("C")},
			{ID: hashD, Data: []byte("D")},
		},
	}

	if err := Save(dir, ss); err != nil {
		t.Fatal(err)
	}

	if err := GC(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hashA[:]))); err != nil {
		t.Fatal("A removed")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hashC[:]))); err != nil {
		t.Fatal("C removed")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hashB[:]))); !os.IsNotExist(err) {
		t.Fatal("B still exists")
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(hashD[:]))); !os.IsNotExist(err) {
		t.Fatal("D still exists")
	}
}

func TestGCIgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()

	hash := testHash("abc")

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID: id,
		Nodes: []Node{
			{
				ID:   1,
				Name: "file",
				Type: TypeFile,
				Chunks: []store.ChunkRef{
					{
						Hash:   hash,
						Length: uint32(len(hash)),
					},
				},
			},
		},
		Objects: []Object{
			{
				ID:   hash,
				Data: []byte("Homu Homu"),
			},
		},
	}

	if err := Save(dir, ss); err != nil {
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
