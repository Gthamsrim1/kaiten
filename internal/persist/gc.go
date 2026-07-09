// Copyright (c) 2026 Gautham Sriram All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package persist

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

func GC(repo string) error {
	snapshotDir := filepath.Join(repo, "snapshots")

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		return err
	}

	live := map[[32]byte]struct{}{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		snap, _, err := LoadSnapshot(filepath.Join(snapshotDir, entry.Name()))
		if err != nil {
			return err
		}

		for _, node := range snap.Nodes {
			for _, chunk := range node.Chunks {
				live[chunk.Hash] = struct{}{}
			}
		}
	}

	objectDir := filepath.Join(repo, "objects")

	entries, err = os.ReadDir(objectDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		hashBytes, err := hex.DecodeString(entry.Name())
		if err != nil {
			continue
		}

		if len(hashBytes) != 32 {
			continue
		}

		var hash [32]byte
		copy(hash[:], hashBytes)

		if _, ok := live[hash]; ok {
			continue
		}

		if err := os.Remove(filepath.Join(objectDir, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}
