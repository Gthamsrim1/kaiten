// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID:     id,
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

	if err := Save(dir, ss); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "snapshots", id+".json")); err != nil {
		t.Fatalf("<id>.json not created: %v", err)
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

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID: id,
		Objects: []Object{
			{
				ID:   name,
				Data: []byte("Madoka"),
			},
		},
	}

	if err := Save(dir, ss); err != nil {
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

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{
		ID:     id,
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

	if err := Save(dir, ss); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "snapshots", id+".json"))
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}

	if _, ok := raw["objects"]; ok {
		t.Fatal("<id>.json should not contain an objects field")
	}
}

func TestSaveEmptySnapshot(t *testing.T) {
	dir := t.TempDir()

	id, err := NewSnapshotID()
	if err != nil {
		t.Fatal(err)
	}

	ss := &Snapshot{ID: id}

	if err := Save(dir, ss); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "snapshots", id+".json")); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "objects")); err != nil {
		t.Fatal(err)
	}
}
