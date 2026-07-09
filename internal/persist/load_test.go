package persist

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	expected := &Snapshot{
		ID: id,
		NextID: 42,
		Nodes: []Node{
			{
				ID:       1,
				ParentID: 0,
				Name:     "/",
				Mode:     040755,
			},
			{
				ID:       2,
				ParentID: 1,
				Name:     "hello",
				Mode:     0100644,
			},
		},
	}

	if err := Save(dir, expected); err != nil {
		t.Fatal(err)
	}

	got, _, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}

	if got.NextID != expected.NextID {
		t.Fatalf("expected NextID %d, got %d", expected.NextID, got.NextID)
	}

	if len(got.Nodes) != len(expected.Nodes) {
		t.Fatalf("expected %d nodes, got %d", len(expected.Nodes), len(got.Nodes))
	}

	for i := range expected.Nodes {
		if !reflect.DeepEqual(got.Nodes[i], expected.Nodes[i]) {
			t.Fatalf(
				"node %d mismatch\nexpected: %#v\ngot: %#v",
				i,
				expected.Nodes[i],
				got.Nodes[i],
			)
		}
	}
}

func TestLoadMissingMetadata(t *testing.T) {
	dir := t.TempDir()

	_, _, err := Load(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadInvalidSnapshot(t *testing.T) {
	dir := t.TempDir()

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	if err := Save(dir, &Snapshot{ID: id}); err != nil {
		t.Fatal(err)
	}

	snapshotPath := filepath.Join(dir, "snapshots", id+".json")

	if err := os.WriteFile(snapshotPath, []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err = Load(dir)
	if err == nil {
		t.Fatal("expected json error")
	}
}
