package persist

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveCreatesFiles(t *testing.T) {
	dir := t.TempDir()

	name := testHash("Kaiten")

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
				ID:   name,
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

	if _, err := os.Stat(filepath.Join(dir, "objects", hex.EncodeToString(name[:]))); err != nil {
		t.Fatalf("object file not created: %v", err)
	}
}

func TestSaveWritesObjectData(t *testing.T) {
	dir := t.TempDir()

	name := testHash("Kaiten")

	fs := &Filesystem{
		Objects: []Object{
			{
				ID:   name,
				Data: []byte("Madoka"),
			},
		},
	}

	if err := Save(dir, fs); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "objects", hex.EncodeToString(name[:])))
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
				ID:   testHash("Kaiten"),
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

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	if _, ok := raw["objects"]; ok {
		t.Fatal("metadata.json should not contain an objects field")
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
