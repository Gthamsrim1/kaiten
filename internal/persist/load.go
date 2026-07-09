package persist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

func Load(path string) (*Snapshot, *Repository, error) {
	head, err := os.ReadFile(filepath.Join(path, "HEAD"))
	if err != nil {
		return nil, nil, err
	}

	snapshotPath := filepath.Join(path, "snapshots", strings.TrimSpace(string(head))+".json")
	return LoadSnapshot(snapshotPath)
}

func LoadSnapshot(snapshotPath string) (*Snapshot, *Repository, error) {
	data, err := os.ReadFile(snapshotPath)
	if err != nil {
		return nil, nil, err
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, nil, err
	}

	repo := &Repository{
		Path: filepath.Dir(filepath.Dir(snapshotPath)),
	}

	return &snap, repo, nil
}
