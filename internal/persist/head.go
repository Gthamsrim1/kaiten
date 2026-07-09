// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package persist

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SnapshotInfo struct {
	ID       string
	ParentID *string
	IsHEAD   bool
}

type snapshotHeader struct {
	ID       string  `json:"id"`
	ParentID *string `json:"parent_id"`
}

func NewSnapshotID() (string, error) {
	var id [16]byte

	if _, err := rand.Read(id[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(id[:]), nil
}

func CurrentSnapshotID(repo string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repo, "HEAD"))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func Checkout(repo, id string) error {
	snapshotPath := filepath.Join(repo, "snapshots", id+".json")

	info, err := os.Stat(snapshotPath)
	switch {
	case os.IsNotExist(err):
		return fmt.Errorf("snapshot %q not found", id)

	case err != nil:
		return err

	case info.IsDir():
		return fmt.Errorf("snapshot %q is a directory", id)
	}

	return writeAtomic(
		filepath.Join(repo, "HEAD"),
		[]byte(id+"\n"),
		0644,
	)
}

func ReadSnapshotInfo(path string) (*SnapshotInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var header snapshotHeader
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}

	return &SnapshotInfo{
		ID:       header.ID,
		ParentID: header.ParentID,
	}, nil
}

func ListSnapshots(repo string) ([]SnapshotInfo, error) {
	head, err := CurrentSnapshotID(repo)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	snapshotDir := filepath.Join(repo, "snapshots")

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return nil, err
	}

	snapshots := make([]SnapshotInfo, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := ReadSnapshotInfo(filepath.Join(snapshotDir, entry.Name()))
		if err != nil {
			return nil, err
		}

		info.IsHEAD = (info.ID == head)

		snapshots = append(snapshots, *info)
	}

	return snapshots, nil
}
