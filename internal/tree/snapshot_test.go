package tree

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
)

func TestSnapshotEmptyFS(t *testing.T) {
	fs := newTestFS()

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if snap.NextID != fs.CurrentID() {
		t.Fatalf("expected NextID %d, got %d", fs.CurrentID(), snap.NextID)
	}

	if len(snap.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(snap.Nodes))
	}

	if len(snap.Objects) != 0 {
		t.Fatalf("expected 0 objects, got %d", len(snap.Objects))
	}
}

func TestSnapshotNodes(t *testing.T) {
	fs := newTestFS()

	_, err := fs.Root.CreateDirectory("bin", 0755)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateDirectory("etc", 0755)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateFile("hello", content.Memory([]byte("hello")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if len(snap.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(snap.Nodes))
	}

	if len(snap.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(snap.Objects))
	}
}

func TestSnapshotParentIDs(t *testing.T) {
	fs := newTestFS()

	dir, err := fs.Root.CreateDirectory("dir", 0755)
	if err != nil {
		t.Fatal(err)
	}

	file, err := dir.CreateFile("file", content.Memory([]byte("hello")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	parents := make(map[uint64]uint64)

	for _, n := range snap.Nodes {
		parents[n.ID] = n.ParentID
	}

	if parents[fs.Root.ID] != 0 {
		t.Fatal("root parent should be 0")
	}

	if parents[dir.ID] != fs.Root.ID {
		t.Fatal("directory parent incorrect")
	}

	if parents[file.ID] != dir.ID {
		t.Fatal("file parent incorrect")
	}
}

func TestSnapshotObjectHash(t *testing.T) {
	fs := newTestFS()

	data := []byte("hello world")

	_, err := fs.Root.CreateFile("hello", content.Memory(data), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if len(snap.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(snap.Objects))
	}

	sum := sha256.Sum256(data)
	expected := hex.EncodeToString(sum[:])

	if snap.Objects[0].ID != expected {
		t.Fatalf("expected hash %q, got %q", expected, snap.Objects[0].ID)
	}
}

func TestSnapshotObjectData(t *testing.T) {
	fs := newTestFS()

	data := []byte("madoka")

	_, err := fs.Root.CreateFile("file", content.Memory(data), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if string(snap.Objects[0].Data) != "madoka" {
		t.Fatalf("expected %q, got %q", "madoka", string(snap.Objects[0].Data))
	}
}

func TestSnapshotDeduplicatesObjects(t *testing.T) {
	fs := newTestFS()

	data := []byte("same")

	_, err := fs.Root.CreateFile("a", content.Memory(data), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateFile("b", content.Memory(data), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if len(snap.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(snap.Nodes))
	}

	if len(snap.Objects) != 1 {
		t.Fatalf("expected deduplicated object store, got %d objects", len(snap.Objects))
	}
}

func TestSnapshotDifferentObjects(t *testing.T) {
	fs := newTestFS()

	_, err := fs.Root.CreateFile("a", content.Memory([]byte("hello")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fs.Root.CreateFile("b", content.Memory([]byte("world")), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	if len(snap.Objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(snap.Objects))
	}
}

func TestSnapshotNodeObjectIDs(t *testing.T) {
	fs := newTestFS()

	data := []byte("hello")

	file, err := fs.Root.CreateFile("file", content.Memory(data), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(data)
	expected := hex.EncodeToString(sum[:])

	for _, n := range snap.Nodes {
		if n.ID != file.ID {
			continue
		}

		if n.ObjectID == nil {
			t.Fatal("file ObjectID should not be nil")
		}

		if *n.ObjectID != expected {
			t.Fatalf("expected %q, got %q", expected, *n.ObjectID)
		}

		return
	}

	t.Fatal("file node not found")
}