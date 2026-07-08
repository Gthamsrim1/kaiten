package persist

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()

	name := testHash("abc123")

	expected := &Filesystem{
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
		Objects: []Object{
			{
				ID:   name,
				Data: []byte("Hello"),
			},
		},
	}

	if err := Save(dir, expected); err != nil {
		t.Fatal(err)
	}

	got, err := Load(dir)
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
		if got.Nodes[i].ID != expected.Nodes[i].ID {
			t.Fatalf("node %d mismatch", i)
		}
	}

	if len(got.Objects) != len(expected.Objects) {
		t.Fatalf("expected %d objects, got %d", len(expected.Objects), len(got.Objects))
	}

	for i := range expected.Objects {
		if got.Objects[i].ID != expected.Objects[i].ID {
			t.Fatalf("expected object id %q, got %q",
				expected.Objects[i].ID,
				got.Objects[i].ID)
		}

		if !bytes.Equal(got.Objects[i].Data, expected.Objects[i].Data) {
			t.Fatalf("object %q data mismatch", expected.Objects[i].ID)
		}
	}
}

func TestLoadMissingMetadata(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadInvalidMetadata(t *testing.T) {
	dir := t.TempDir()

	if err := Save(dir, &Filesystem{}); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), []byte("{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected json error")
	}
}

func TestLoadMissingObject(t *testing.T) {
	dir := t.TempDir()
	name := testHash("Homura")

	fs := &Filesystem{
		Objects: []Object{
			{
				ID:   name,
				Data: []byte("Ai yo"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(dir, "objects", hex.EncodeToString(name[:]))); err != nil {
		t.Fatal(err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for missing object")
	}
}
