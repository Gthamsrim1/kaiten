package persist

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

func GC(repo string) error {
	fs, _, err := Load(repo)
	if err != nil {
		return err
	}

	live := map[[32]byte]struct{}{}

	for _, node := range fs.Nodes {
		for _, chunk := range node.Chunks {
			live[chunk.Hash] = struct{}{}
		}
	}

	objectDir := filepath.Join(repo, "objects")

	entries, err := os.ReadDir(objectDir)
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
