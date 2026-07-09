package tree

import (
	"testing"

	"github.com/Gthamsrim1/kaiten/internal/content"
	"github.com/Gthamsrim1/kaiten/internal/persist"
)

func snapshotTestFS(t *testing.T, fs *KaitenFS) *persist.Snapshot {
	t.Helper()

	id, err := persist.NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	snap, err := fs.Snapshot(id, nil)
	if err != nil {
		t.Fatal(err)
	}

	return snap
}

func TestSnapshotEmptyFS(t *testing.T) {
	fs := newTestFS()

	snap := snapshotTestFS(t, fs)

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

	snap := snapshotTestFS(t, fs)

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

	snap := snapshotTestFS(t, fs)

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

	snap := snapshotTestFS(t, fs)

	if len(snap.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(snap.Objects))
	}

	if len(snap.Nodes[1].Chunks) != 1 {
		t.Fatal("expected one chunk")
	}

	if snap.Objects[0].ID != snap.Nodes[1].Chunks[0].Hash {
		t.Fatal("object hash does not match chunk reference")
	}
}

func TestSnapshotObjectData(t *testing.T) {
	fs := newTestFS()

	data := []byte("madoka")

	_, err := fs.Root.CreateFile("file", content.Memory(data), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap := snapshotTestFS(t, fs)

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

	snap := snapshotTestFS(t, fs)

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

	snap := snapshotTestFS(t, fs)

	if len(snap.Objects) != 2 {
		t.Fatalf("expected 2 objects, got %d", len(snap.Objects))
	}
}

func TestSnapshotNodeChunks(t *testing.T) {
	fs := newTestFS()

	data := []byte("hello")

	file, err := fs.Root.CreateFile("file", content.Memory(data), 0644)
	if err != nil {
		t.Fatal(err)
	}

	snap := snapshotTestFS(t, fs)

	for _, n := range snap.Nodes {
		if n.ID != file.ID {
			continue
		}

		if len(n.Chunks) != 1 {
			t.Fatalf("expected 1 chunk, got %d", len(n.Chunks))
		}

		if n.Chunks[0].Length != uint32(len(data)) {
			t.Fatal("incorrect chunk length")
		}

		if n.Chunks[0].Hash != snap.Objects[0].ID {
			t.Fatal("chunk hash does not match object")
		}

		return
	}

	t.Fatal("file node not found")
}

func TestSnapshotID(t *testing.T) {
	fs := newTestFS()

	snap := snapshotTestFS(t, fs)

	if snap.ID == "" {
		t.Fatal("expected snapshot ID")
	}
}

func TestSnapshotParentID(t *testing.T) {
	fs := newTestFS()

	id, err := persist.NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	parent := "parent-snapshot"

	snap, err := fs.Snapshot(id, &parent)
	if err != nil {
		t.Fatal(err)
	}

	if snap.ParentID == nil {
		t.Fatal("expected parent")
	}

	if *snap.ParentID != parent {
		t.Fatalf("expected %q, got %q", parent, *snap.ParentID)
	}
}
