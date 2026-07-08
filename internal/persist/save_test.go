package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveCreatesFiles(t *testing.T) {
	dir := t.TempDir()

	fs := &Filesystem{
		NextID: 2,
		Nodes: []Node{
			{
				ID:   1,
				Name: "/",
			},
		},
		Objects: []Object{
			{
				ID:   "abc123",
				Data: []byte("Homura"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
		t.Fatalf("metadata.json not created: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects")); err != nil {
		t.Fatalf("objects directory not created: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects", "abc123")); err != nil {
		t.Fatalf("object file not created: %v", err)
	}
}

func TestSaveWritesObjectData(t *testing.T) {
	dir := t.TempDir()

	fs := &Filesystem{
		Objects: []Object{
			{
				ID:   "object1",
				Data: []byte("Madoka"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "objects", "object1"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "Madoka" {
		t.Fatalf("expected %q, got %q", "Madoka", string(data))
	}
}

func TestSaveMetadataDoesNotContainObjects(t *testing.T) {
	dir := t.TempDir()

	fs := &Filesystem{
		NextID: 2,
		Nodes: []Node{
			{
				ID:   1,
				Name: "/",
			},
		},
		Objects: []Object{
			{
				ID:   "Kaiten",
				Data: []byte("Madoka"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "metadata.json"))
	if err != nil {
		t.Fatal(err)
	}

	var meta Filesystem
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatal(err)
	}

	if len(meta.Objects) != 1 {
		t.Fatalf("expected 1 object reference, got %d", len(meta.Objects))
	}

	if meta.Objects[0].ID != "Kaiten" {
		t.Fatalf("unexpected object id %q", meta.Objects[0].ID)
	}
}

func TestSaveEmptyFilesystem(t *testing.T) {
	dir := t.TempDir()

	fs := &Filesystem{}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects")); err != nil {
		t.Fatal(err)
	}
}